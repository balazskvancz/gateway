package gateway

import (
	"log"
	"os"
)

type gatewayLogger struct {
	*log.Logger
}

type logger interface {
	info(string)
	error(string)
	warning(string)
}

var _ logger = (*gatewayLogger)(nil)

func newGatewayLogger() logger {
	return &gatewayLogger{
		Logger: log.New(os.Stdout, "api-gateway", 0),
	}
}

func (l *gatewayLogger) info(v string) {}

func (l *gatewayLogger) error(v string) {}

func (l *gatewayLogger) warning(v string) {}
