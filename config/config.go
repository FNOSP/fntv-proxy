package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	Port int `ini:"port"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	// 检查配置文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 创建默认配置文件
		err = createDefaultConfig(filename)
		if err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		log.Printf("已创建默认配置文件: %s", filename)
	}

	// 加载配置文件
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	port, err := cfg.Section("server").Key("port").Int()
	if err != nil {
		log.Printf("配置文件中端口格式不正确，使用默认端口1999")
		port = 1999
	}

	config.Port = port
	return config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(filename string) error {
	cfg := ini.Empty()
	cfg.Section("server").Key("port").SetValue("1999")
	return cfg.SaveTo(filename)
}
