package redis_adapter

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
	"time"
)

var _ persist.Adapter = (*ClusterAdapter)(nil)

var _ persist.SerializerAdapter = (*ClusterAdapter)(nil)

var _ persist.BatchAdapter = (*ClusterAdapter)(nil)

type ClusterAdapter struct {
	client *redis.ClusterClient
}

func (r *ClusterAdapter) GetClient() *redis.ClusterClient {
	return r.client
}

func (r *ClusterAdapter) Serialize(session *model.Session) ([]byte, error) {
	return json.Marshal(session)
}

func (r *ClusterAdapter) UnSerialize(bytes []byte) (*model.Session, error) {
	s := &model.Session{}
	err := json.Unmarshal(bytes, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func NewClusterAdapter(addrs []string, username string, password string) *ClusterAdapter {
	return NewClusterAdapterByOptions(
		&redis.ClusterOptions{
			Addrs:    addrs,
			Username: username,
			Password: password,
		})
}

func NewClusterAdapterByOptions(clusterOptions *redis.ClusterOptions) *ClusterAdapter {
	client := redis.NewClusterClient(clusterOptions)
	return &ClusterAdapter{client: client}
}

func (r *ClusterAdapter) GetStr(key string) string {
	res, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	return res
}

func (r *ClusterAdapter) SetStr(key string, value string, timeout int64) error {
	err := r.client.Set(context.Background(), key, value, time.Duration(timeout)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) UpdateStr(key string, value string) error {
	err := r.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) DeleteStr(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) GetStrTimeout(key string) int64 {
	duration, err := r.client.TTL(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	return int64(duration.Seconds())
}

func (r *ClusterAdapter) UpdateStrTimeout(key string, timeout int64) error {
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

func (r *ClusterAdapter) Get(key string) interface{} {
	res, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}
	s := &model.Session{}
	err = json.Unmarshal([]byte(res), s)
	if err != nil {
		return nil
	}
	return s
}

func (r *ClusterAdapter) Set(key string, value interface{}, timeout int64) error {
	err := r.client.Set(context.Background(), key, value, time.Duration(timeout)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) Update(key string, value interface{}) error {
	err := r.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) Delete(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterAdapter) GetTimeout(key string) int64 {
	duration, err := r.client.TTL(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	return int64(duration.Seconds())
}

func (r *ClusterAdapter) UpdateTimeout(key string, timeout int64) error {
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

func (r *ClusterAdapter) DeleteBatchFilteredKey(filterKeyPrefix string) error {
	err := r.client.ForEachMaster(context.Background(), func(ctx context.Context, client *redis.Client) error {
		var cursor uint64
		for {
			keys, cursor, err := client.Scan(context.Background(), cursor, filterKeyPrefix+"*", 100).Result()
			if err != nil {
				return err
			}

			if len(keys) == 0 && cursor == 0 {
				break
			}

			// use pip delete batch
			pipe := client.Pipeline()

			for _, key := range keys {
				pipe.Del(context.Background(), key)
			}

			_, err = pipe.Exec(context.Background())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
