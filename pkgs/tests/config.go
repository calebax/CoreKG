package tests

import (
	"flag"
	"fmt"

	"github.com/ygpkg/yg-go/config"
)

var testCfg *config.CoreConfig

// InitConfig
func InitConfig() (*config.CoreConfig, error) {
	if testCfg != nil {
		return testCfg, nil
	}
	if !flag.Parsed() {
		flag.Parse()
	}
	argList := flag.Args()
	configFile := ""
	if len(argList) == 1 {
		configFile = argList[0]
	}
	if configFile == "" {
		return nil, fmt.Errorf("config file is empty")
	}
	cfg, err := config.LoadCoreConfig(configFile)
	if err != nil {
		return nil, err
	}
	testCfg = cfg
	return cfg, nil
}
