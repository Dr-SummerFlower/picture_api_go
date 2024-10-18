package config

import (
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"summerflower.local/picture_api/server/logger"
)

type config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
}

var Config config

func init() {
	file, err := os.Open("resources/config.yaml")
	if err != nil {
		logger.Log.Error("读取配置文件出错", zap.Error(err))
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Error("无法关闭文件", zap.Error(err))
			return
		}
	}(file)

	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&Config)
	if err != nil {
		logger.Log.Error("解析配置文件出错", zap.Error(err))
		return
	}
}
