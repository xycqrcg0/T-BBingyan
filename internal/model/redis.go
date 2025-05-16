package model

import (
	"BBingyan/internal/config"
	"BBingyan/internal/log"
	"github.com/go-redis/redis"
)

var RedisDB *redis.Client

func newRedis() {
	RedisDB = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		DB:       0,
		Password: "",
	})
	var err error
	for i := 0; i < 10; i++ {
		if _, err = RedisDB.Ping().Result(); err != nil {
			log.Warnf("%d fail to connect to redis", i)
		} else {
			break
		}
	}
	if err != nil {
		log.Fatalf("Fail to init redis")
	}
	log.Infof("Finish initializing redis")
}
