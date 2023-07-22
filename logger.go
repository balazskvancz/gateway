package gateway

import (
	"fmt"
	"log"
	"os"
)

type gatewayLogger struct {
	*log.Logger
}

type logger interface {
	info([]byte)
	error(string)
	warning(string)
}

const (
	defaultLogFlag = log.LstdFlags
)

var _ logger = (*gatewayLogger)(nil)

func newGatewayLogger() logger {
	logPrefix := fmt.Sprintf("[api-gateway %s] ", Version)
	return &gatewayLogger{
		Logger: log.New(os.Stdout, logPrefix, defaultLogFlag),
	}
}

func (l *gatewayLogger) info(b []byte) {
	l.Printf(string(b))
	// l.Writer().Write(b)
}

func (l *gatewayLogger) error(v string) {}

func (l *gatewayLogger) warning(v string) {}
