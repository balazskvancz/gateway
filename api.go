package gateway

import (
	"reflect"
)

type ServiceInfo struct {
	*ServiceConfig
	State string `json:"state"`
}

type infoResponse struct {
	TotalConnectionService uint64         `json:"totalConnectionServed"`
	IsProd                 bool           `json:"isProd"`
	AreMiddlewaresEnabled  bool           `json:"areMiddlewaresEnabled"`
	Uptime                 string         `json:"uptime"`
	Services               []*ServiceInfo `json:"services"`
}

type updateServiceStateRequest struct {
	ServiceName string `json:"serviceName"`
}

const (
	IncomingDecodedKey contextKey = "incomingDecoded"
	X_GW_HEADER_KEY    string     = "X-GATEWAY-KEY"
)

type decodeFunction func([]byte) (any, error)

// validateIncomingRequest validates all the incoming requests by its header key.
func validateIncomingRequest(g *Gateway, df decodeFunction) MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) {
		b, err := ctx.GetRawBody()
		if err != nil {
			ctx.SendUnauthorized()
			return
		}

		var (
			key   = ctx.GetRequestHeader(X_GW_HEADER_KEY)
			plain = append(b, []byte(g.info.secretKey)...)
		)

		if h := createHash(plain); reflect.DeepEqual(h, []byte(key)) {
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

// serviceStateUpdateHandler returns a HandlerFunc which will update the corresponding service's state.
func serviceStateUpdateHandler(g *Gateway) HandlerFunc {
	return func(ctx *Context) {
		inc, err := getValueFromContext[*updateServiceStateRequest](ctx.ctx, IncomingDecodedKey)
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

// getSystemInfoHandler returns a response with the Gateway's info.
// Currently it only returns the slice of registered services – with all its info –
// the system's uptime and the count of served connections so far.
func getSystemInfoHandler(g *Gateway) HandlerFunc {
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
			IsProd:                 g.isProd(),
			AreMiddlewaresEnabled:  g.areMiddlewaresEnabled(),
			Uptime:                 getElapsedTime(g.info.startTime),
		}

		ctx.SendJson(res)
	}
}
