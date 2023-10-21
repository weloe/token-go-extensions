package redis_updatablewatcher

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type WatcherOptions struct {
	Options        redis.Options
	ClusterOptions redis.ClusterOptions
	SubClient      *redis.Client
	PubClient      *redis.Client
	Channel        string
	IgnoreSelf     bool
	LocalID        string
}

func initConfig(option *WatcherOptions) {
	if option.LocalID == "" {
		option.LocalID = uuid.New().String()
	}
	if option.Channel == "" {
		option.Channel = "/token-go"
	}
}
