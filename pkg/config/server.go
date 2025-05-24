package config

import "os"

type ServerConfig struct {
	Port           string `json:"port"`
	ClickHouseAddr string `json:"clickhouse_addr"`
	ClickHouseDB   string `json:"clickhouse_db"`
	ClickHouseUser string `json:"clickhouse_user"`
	ClickHousePass string `json:"clickhouse_pass"`
}

func LoadServerConfig() ServerConfig {
	config := ServerConfig{
		Port:           ":8080",
		ClickHouseAddr: "clickhousedb.axe.steigerlabs.com:9000",
		ClickHouseDB:   "vlgpus",
		ClickHouseUser: "default",
		ClickHousePass: "",
	}

	if port := os.Getenv("PORT"); port != "" {
		config.Port = ":" + port
	}
	if addr := os.Getenv("CLICKHOUSE_ADDR"); addr != "" {
		config.ClickHouseAddr = addr
	}
	if db := os.Getenv("CLICKHOUSE_DB"); db != "" {
		config.ClickHouseDB = db
	}
	if user := os.Getenv("CLICKHOUSE_USER"); user != "" {
		config.ClickHouseUser = user
	}
	if pass := os.Getenv("CLICKHOUSE_PASS"); pass != "" {
		config.ClickHousePass = pass
	}

	return config
}