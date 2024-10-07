package config

import (
	"time"

	"github.com/spf13/viper"
)

type Postgres struct {
	Port     int64
	Host     string
	User     string
	Password string
	DBName   string
}

type Config struct {
	Port             int64
	Host             string
	ZerosCount       int64
	RandThreshold    int64
	ClientMaxCounter int64
	TTL              time.Duration
	ClientDelay      time.Duration
	Postgres         *Postgres
	Quotes           []string
}

func ParseConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err = viper.Unmarshal(c); err != nil {
		return nil, err
	}

	return c, nil
}
