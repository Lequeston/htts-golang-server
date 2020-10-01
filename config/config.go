package config

import "os"

type Config struct {
	PortHTTP int
	PortHTTPS int
}

var Production = Config{
	PortHTTP: 80,
	PortHTTPS: 443,
}

var Development = Config{
	PortHTTP: 4041,
	PortHTTPS: 4040,
}

func GetConfigValues() *Config {
	isProd := os.Getenv("NODE_ENV") == "production"
	if isProd {
		return &Production
	} else {
		return &Development
	}
}
