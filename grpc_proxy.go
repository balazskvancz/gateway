// This lightweight gRPC proxy was based on the work:
// https://github.com/mwitkow/grpc-proxy
// I am grateful for publishing that package, helped me a lot!
// Thanks for the authors!
package gateway

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	proxyDesc = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

type serviceLookupFn func(string) *service

type grpcProxy struct {
	logger
	address       int
	stopChan      chan struct{}
	server        *grpc.Server
	serviceLookup serviceLookupFn
}

func newGrpcProxy(address int, l logger, fn serviceLookupFn) *grpcProxy {
	proxy := &grpcProxy{
		logger:        l,
		address:       address,
		stopChan:      make(chan struct{}),
		serviceLookup: fn,
	}

	proxy.server = grpc.NewServer(
		grpc.UnknownServiceHandler(proxy.handler),
	)

	return proxy
}

func (g *grpcProxy) listen() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", g.address))
	if err != nil {

		return
	}

	go func() {
		if err := g.server.Serve(ln); err != nil {
			fmt.Println(err)
		}
	}()
	<-g.stopChan
	g.server.Stop()
}

func (g *grpcProxy) stop() {
	close(g.stopChan)
}

func (g *grpcProxy) handler(srv interface{}, serverStream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}

	var (
		serviceName = getServiceNameByGRPCMethod(fullMethodName)
		service     = g.serviceLookup(serviceName)
	)

	if service == nil {
		return status.Errorf(codes.Internal, "service %s not found", serviceName)
	}

	conn, err := grpc.Dial(service.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	ctx := context.TODO() // Change it to the associated service's context with timeout.

	clientStream, err := grpc.NewClientStream(ctx, proxyDesc, conn, fullMethodName)
	if err != nil {
		return err
	}
	var (
		s2cErrChan = forwardServerToClient(serverStream, clientStream)
		c2sErrChan = forwardClientToServer(clientStream, serverStream)
	)
	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				clientStream.CloseSend()
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}

func forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	var (
		ret          = make(chan error, 1)
		isHeaderRead bool
	)
	go func() {
		f := &emptypb.Empty{}
		for {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if !isHeaderRead {
				isHeaderRead = true
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely.
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func getServiceNameByGRPCMethod(m string) string {
	// /example.TestService/...
	spl := strings.Split(m, "/")
	if len(spl) < 2 {
		return ""
	}
	return "/" + spl[1]
}
