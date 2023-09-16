package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultConfigPath = "./config.json"
)

type LoggerConfig struct {
	DisabledLoggers []logTypeName `json:"disabledLoggers"`
}

type GrpcProxyConfig struct {
	Address int `json:"address"`
}

type GatewayConfig struct {
	Address             int              `json:"address"`
	MiddlewaresEnabled  *runLevel        `json:"middlewaresEnabled"`
	ProductionLevel     *runLevel        `json:"productionLevel"`
	SecretKey           string           `json:"secretKey"`
	HealthCheckInterval string           `json:"healthCheckInterval"`
	TimeOutSec          int              `json:"timeOutSec"`
	LoggerConfig        *LoggerConfig    `json:"loggerConfig"`
	GrpcProxy           *GrpcProxyConfig `json:"grpcProxy"`

	Services []*ServiceConfig `json:"services"`
}

type duration byte

const (
	durationSecond duration = 's'
	durationMinute duration = 'm'
)

func getHealthCheckInterval(t string) time.Duration {
	if t == "" {
		return 0
	}

	var (
		val = t[:len(t)-1]
		d   = t[len(t)-1]
	)

	timeValue, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	if d == byte(durationMinute) {
		return time.Minute * time.Duration(timeValue)
	}

	if d == byte(durationSecond) {
		return time.Second * time.Duration(timeValue)
	}

	return 0
}

// ReadConfig reads the JSON config from given path,
// then returns it as a slice of GatewayOptionFunc,
// which can be passed into the New factory.
// In case of unexpected behaviour, it returns error.
func ReadConfig(path string) ([]GatewayOptionFunc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf, err := parseConfig(b)
	if err != nil {
		return nil, err
	}

	funcs := getGatewayOptionFuncs(conf)

	return funcs, nil
}

func getGatewayOptionFuncs(conf *GatewayConfig) []GatewayOptionFunc {
	funcs := make([]GatewayOptionFunc, 0)

	if conf.Address > 0 {
		funcs = append(funcs, WithAddress(conf.Address))
	}

	if conf.SecretKey != "" {
		funcs = append(funcs, WithSecretKey(conf.SecretKey))
	}

	for _, conf := range conf.Services {
		funcs = append(funcs, WithService(conf))
	}

	if conf.MiddlewaresEnabled != nil {
		funcs = append(funcs, WithMiddlewaresEnabled(*conf.MiddlewaresEnabled))
	}

	if conf.ProductionLevel != nil {
		funcs = append(funcs, WithProductionLevel(*conf.ProductionLevel))
	}

	if conf.GrpcProxy != nil {
		funcs = append(funcs, WithGrpcProxy(conf.GrpcProxy.Address))
	}

	if configInterval := getHealthCheckInterval(conf.HealthCheckInterval); configInterval != 0 {
		funcs = append(funcs, WithHealthCheckFrequency(configInterval))
	}

	if conf.LoggerConfig != nil {
		var value logTypeValue = 0
		for _, e := range conf.LoggerConfig.DisabledLoggers {
			val, ok := logLevelValues[e]
			if !ok {
				continue
			}
			value += val
		}
		funcs = append(funcs, WithDisabledLoggers(value))
	}

	return funcs
}

func parseConfig(b []byte) (*GatewayConfig, error) {
	conf := &GatewayConfig{}
	if err := json.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
