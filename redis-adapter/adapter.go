package redis_adapter

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/weloe/token-go/persist"
	"github.com/weloe/token-go/util"
	"log"
	"reflect"
	"time"
)

var _ persist.Adapter = (*RedisAdapter)(nil)

type RedisAdapter struct {
	client     *redis.Client
	serializer persist.Serializer
}

func (r *RedisAdapter) SetSerializer(serializer persist.Serializer) {
	r.serializer = serializer
}

func (r *RedisAdapter) GetClient() *redis.Client {
	return r.client
}

func NewAdapter(addr string, username string, password string, db int) (*RedisAdapter, error) {
	return NewAdapterByOptions(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})
}

func NewAdapterByOptions(options *redis.Options) (*RedisAdapter, error) {
	client := redis.NewClient(options)
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return &RedisAdapter{client: client, serializer: persist.NewJsonSerializer()}, nil
}

func (r *RedisAdapter) GetStr(key string) string {
	res, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	return res
}

func (r *RedisAdapter) SetStr(key string, value string, timeout int64) error {
	err := r.client.Set(context.Background(), key, value, time.Duration(timeout)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) UpdateStr(key string, value string) error {
	err := r.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) DeleteStr(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) GetStrTimeout(key string) int64 {
	duration, err := r.client.TTL(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	return int64(duration.Seconds())
}

func (r *RedisAdapter) UpdateStrTimeout(key string, timeout int64) error {
	var duration time.Duration
	if timeout < 0 {
		duration = -1
	} else {
		duration = time.Duration(timeout) * time.Second
	}
	err := r.client.Expire(context.Background(), key, duration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) Get(key string, t ...reflect.Type) interface{} {
	value, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	if r.serializer == nil || t == nil || len(t) == 0 {
		return value
	}
	bytes, err := util.InterfaceToBytes(value)
	if err != nil {
		log.Printf("Adapter.Get() failed: %v", err)
		return nil
	}
	instance := reflect.New(t[0].Elem()).Interface()
	err = r.serializer.UnSerialize(bytes, instance)
	if err != nil {
		log.Printf("Adapter.Get() failed: %v", err)
		return nil
	}

	return instance
}

func (r *RedisAdapter) Set(key string, value interface{}, timeout int64) error {
	var err error
	if r.serializer != nil {
		bytes, err := r.serializer.Serialize(value)
		if err != nil {
			return err
		}
		err = r.client.Set(context.Background(), key, bytes, time.Duration(timeout)*time.Second).Err()
	} else {
		err = r.client.Set(context.Background(), key, value, time.Duration(timeout)*time.Second).Err()
	}

	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) Update(key string, value interface{}) error {
	var err error
	if r.serializer != nil {
		bytes, err := r.serializer.Serialize(value)
		if err != nil {
			return err
		}
		err = r.client.Set(context.Background(), key, bytes, 0).Err()
	} else {
		err = r.client.Set(context.Background(), key, value, 0).Err()
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) Delete(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) GetTimeout(key string) int64 {
	duration, err := r.client.TTL(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	return int64(duration.Seconds())
}

func (r *RedisAdapter) UpdateTimeout(key string, timeout int64) error {
	var duration time.Duration
	if timeout < 0 {
		duration = -1
	} else {
		duration = time.Duration(timeout) * time.Second
	}
	err := r.client.Expire(context.Background(), key, duration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisAdapter) DeleteBatchFilteredKey(filterKeyPrefix string) error {
	var cursor uint64
	for {
		keys, cursor, err := r.client.Scan(context.Background(), cursor, filterKeyPrefix+"*", 100).Result()
		if err != nil {
			return err
		}

		if len(keys) == 0 && cursor == 0 {
			break
		}

		// use pip delete batch
		pipe := r.client.Pipeline()

		for _, key := range keys {
			pipe.Del(context.Background(), key)
		}

		_, err = pipe.Exec(context.Background())
		if err != nil {
			return err
		}
	}

	return nil
}
