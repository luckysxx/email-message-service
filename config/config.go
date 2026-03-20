package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig   `mapstructure:"app"`
	Kafka KafkaConfig `mapstructure:"kafka"`
	SMTP  SMTPConfig  `mapstructure:"smtp"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// LoadConfig 从指定路径加载配置文件并解析到 Config 结构体中
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)

	// 提供默认值
	viper.SetDefault("app.env", "development")

	// 读取文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	// 允许通过环境变量覆盖配置，例如设置 APP_ENV
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}

	return &cfg, nil
}
