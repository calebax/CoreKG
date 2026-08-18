package systemconfig

import "github.com/ygpkg/yg-go/settings"

type SystemConfig struct {
	AlgoModelID uint `yaml:"algo_model_id"` // 算法模型ID
}

func GetSystemConfig() (*SystemConfig, error) {
	config := &SystemConfig{}
	err := settings.GetYaml("knowledge", "system_config", config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
