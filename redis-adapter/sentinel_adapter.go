package redis_adapter

import (
	"github.com/go-redis/redis/v8"
	"github.com/weloe/token-go/persist"
)

var _ persist.Adapter = (*SentinelAdapter)(nil)

var _ persist.SerializerAdapter = (*SentinelAdapter)(nil)

var _ persist.BatchAdapter = (*SentinelAdapter)(nil)

type SentinelAdapter struct {
	*RedisAdapter
}

// NewSentinelAdapter adapter for sentinel mode
func NewSentinelAdapter(masterName string, addrs []string, username string, password string, db int) *SentinelAdapter {
	return NewSentinelAdapterByOptions(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: addrs,
		Username:      username,
		Password:      password,
		DB:            db,
	})
}

func NewSentinelAdapterByOptions(options *redis.FailoverOptions) *SentinelAdapter {
	return &SentinelAdapter{&RedisAdapter{client: redis.NewFailoverClient(options)}}
}
