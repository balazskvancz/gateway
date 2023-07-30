package gateway

import (
	"encoding/json"
	"os"
)

const (
	defaultConfigPath = "./config.json"
)

type GatewayConfig struct {
	Address            int       `json:"address"`
	MiddlewaresEnabled *runLevel `json:"middlewaresEnabled"`
	ProductionLevel    *runLevel `json:"productionLevel"`
	SecretKey          string    `json:"secretKey"`

	Services []*ServiceConfig `json:"services"`
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

	return funcs, nil
}

func parseConfig(b []byte) (*GatewayConfig, error) {
	conf := &GatewayConfig{}
	if err := json.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
