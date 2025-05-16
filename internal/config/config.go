package config

import (
	"BBingyan/internal/log"
	"encoding/json"
	"github.com/joho/godotenv"
	"os"
)

type PostgresConfig struct {
	Dsn string `json:"dsn"`
}

type RedisConfig struct {
	Addr string `json:"addr"`
}

type AdminConfig struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type JwtConfig struct {
	Key     string   `json:"key"`
	Exp     int      `json:"exp"`
	Skipper []string `json:"skipper"`
	User    []string `json:"user"`  //只有user用户能用的接口
	Admin   []string `json:"admin"` //只有admin用户能用的接口
}

type CurdConfig struct {
	PageSize int      `json:"page-size"`
	Tags     []string `json:"tags"`
}

type StructConfig struct {
	AuthorizationCode string
	Port              string         `json:"port"`
	Postgres          PostgresConfig `json:"postgres"`
	Redis             RedisConfig    `json:"redis"`
	Admin             AdminConfig    `json:"admin"`
	JWT               JwtConfig      `json:"jwt"`
	Curd              CurdConfig     `json:"curd"`
}

var Config StructConfig

//ps:admin暂且没有用（没写admin账户）

func InitConfig() {
	file, err := os.ReadFile("./config/config.json")
	if err != nil {
		log.Fatalf("Fail to read from cinfig.json")
	}
	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalf("Fail to unmarshal config.json")
	}
	log.Infof("finish initializing config")

	_ = godotenv.Load()
	Config.AuthorizationCode = os.Getenv("AUTH_CODE")
	log.Warnf("exp:%d", Config.JWT.Exp)

	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		Config.Redis.Addr = addr
	}
	if dsn := os.Getenv("PG_DSN"); dsn != "" {
		Config.Postgres.Dsn = dsn
	}
}
