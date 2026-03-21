package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
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

// LoadConfig 从 Viper 加载配置，支持 .env 文件覆盖机密数据
func LoadConfig() *Config {
	// 尝试从根目录读取 .env 文件
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/app")

	// 允许环境变量覆盖配置 (比如 SMTP_PASSWORD=xxx)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("警告: 未找到 config.yaml，将完全依赖环境变量: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("配置反序列化失败: %v", err)
	}

	return &cfg
}
