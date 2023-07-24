package gateway

import (
	"context"
	"fmt"
)

type ServiceInfo struct {
	*ServiceConfig
	State string `json:"state"`
}

type infoResponse struct {
	TotalConnectionService uint64         `json:"totalConnectionServed"`
	Services               []*ServiceInfo `json:"services"`
}

type updateServiceStateRequest struct {
	ServiceName string `json:"serviceName"`
}

const (
	IncomingDecodedKey = "incomingDecoded"
	X_GW_HEADER_KEY    = "X-GATEWAY-KEY"
)

type decodeFunction func([]byte) (any, error)

func validateIncomingRequest(g *Gateway, df decodeFunction) MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) {
		b, err := ctx.GetRawBody()
		if err != nil {
			ctx.SendUnauthorized()
			return
		}

		var (
			key   = ctx.GetRequestHeader(X_GW_HEADER_KEY)
			plain = append(b, []byte(g.info.SecretKey)...)
		)

		if h := createHash(plain); h != key {
			fmt.Println(h)
			fmt.Println("nincs kulcs xd.")
			ctx.SendUnauthorized()
			return
		}

		incoming, err := df(b)
		if err != nil {
			ctx.SendUnauthorized()
			return
		}

		ctx.BindValue(IncomingDecodedKey, incoming)
		next(ctx)
	}
}

func serviceStateUpdateHandler(g *Gateway) HandlerFunc {
	return func(ctx *Context) {
		inc, err := getFromContext[*updateServiceStateRequest](ctx.ctx, IncomingDecodedKey)
		if err != nil {
			ctx.SendUnauthorized()
			return
		}

		if inc.ServiceName == "" {
			ctx.SendUnauthorized()
			return
		}

		g.serviceRegisty.setServiceAvailable(inc.ServiceName)
		ctx.SendOk()
	}
}

func getFromContext[T any](ctx context.Context, key string) (T, error) {
	val, ok := ctx.Value(key).(T)
	if !ok {
		var def T
		return def, fmt.Errorf("cant parse from key: %s", key)
	}
	return val, nil
}

func getServiceStateHandler(g *Gateway) HandlerFunc {
	return func(ctx *Context) {
		services := g.serviceRegisty.getAllServices()

		info := make([]*ServiceInfo, len(services))

		for i, e := range services {
			info[i] = &ServiceInfo{
				ServiceConfig: e.ServiceConfig,
				State:         stateTexts[e.state],
			}
		}

		res := &infoResponse{
			TotalConnectionService: ctx.contextId,
			Services:               info,
		}

		ctx.SendJson(res)
	}
}
