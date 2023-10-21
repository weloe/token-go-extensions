package redis_updatablewatcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	tokengo "github.com/weloe/token-go"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
	"log"
	"strings"
	"sync"
)

func DefaultUpdateCallback(e *tokengo.Enforcer) func(msg *MSG) {
	return func(msg *MSG) {
		var err error
		adapter := e.GetAdapter()
		switch msg.Method {
		case UpdateForSetStr:
			err = adapter.SetStr(msg.K, msg.V.(string), msg.Timeout)
		case UpdateForUpdateStr:
			err = adapter.UpdateStr(msg.K, msg.V.(string))
		case UpdateForSetSession:
			err = adapter.Set(msg.K, msg.V, msg.Timeout)
		case UpdateForUpdateSession:
			err = adapter.Update(msg.K, msg.V)
		case UpdateForSetQRCode:
			err = adapter.Set(msg.K, msg.V, msg.Timeout)
		case UpdateForUpdateQRCode:
			err = adapter.Update(msg.K, msg.V)
		case UpdateForDelete:
			err = adapter.Delete(msg.K)
		case UpdateForUpdateTimeout:
			err = adapter.UpdateTimeout(msg.K, msg.Timeout)
		default:
			err = errors.New("unknown update type")
		}
		if err != nil {
			log.Println(err)
			log.Println("callback update failed")
		}
	}
}

var _ persist.UpdatableWatcher = &Watcher{}

type Watcher struct {
	client    *redis.Client
	ctx       context.Context
	l         sync.Mutex
	subClient redis.UniversalClient
	pubClient redis.UniversalClient
	options   WatcherOptions
	close     chan struct{}
	callback  func(*MSG)
}

func NewWatcherWithCluster(addrs string, option WatcherOptions) (persist.UpdatableWatcher, error) {
	addrsStr := strings.Split(addrs, ",")
	option.ClusterOptions.Addrs = addrsStr
	initConfig(&option)

	w := &Watcher{
		subClient: redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrsStr,
			Password: option.ClusterOptions.Password,
		}),
		pubClient: redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrsStr,
			Password: option.ClusterOptions.Password,
		}),
		ctx:   context.Background(),
		close: make(chan struct{}),
	}

	err := w.initConfig(option, true)
	if err != nil {
		return nil, err
	}

	if err := w.subClient.Ping(w.ctx).Err(); err != nil {
		return nil, err
	}
	if err := w.pubClient.Ping(w.ctx).Err(); err != nil {
		return nil, err
	}

	w.options = option

	w.subscribe()

	return w, nil
}

func NewWatcher(addr string, option WatcherOptions) (persist.UpdatableWatcher, error) {
	option.Options.Addr = addr
	initConfig(&option)
	w := &Watcher{
		ctx:   context.Background(),
		close: make(chan struct{}),
	}

	if err := w.initConfig(option); err != nil {
		return nil, err
	}

	if err := w.subClient.Ping(w.ctx).Err(); err != nil {
		return nil, err
	}
	if err := w.pubClient.Ping(w.ctx).Err(); err != nil {
		return nil, err
	}

	w.options = option

	w.subscribe()

	return w, nil
}

func (w *Watcher) SetUpdateCallback(callback func(*MSG)) error {
	w.l.Lock()
	w.callback = callback
	w.l.Unlock()
	return nil
}

func (w *Watcher) GetWatcherOptions() WatcherOptions {
	w.l.Lock()
	defer w.l.Unlock()
	return w.options
}

func (w *Watcher) unsubscribe(psc *redis.PubSub) error {
	return psc.Unsubscribe(w.ctx)
}

func (w *Watcher) subscribe() {
	w.l.Lock()
	sub := w.subClient.Subscribe(w.ctx, w.options.Channel)
	w.l.Unlock()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			err := sub.Close()
			if err != nil {
				log.Println(err)
			}
			err = w.pubClient.Close()
			if err != nil {
				log.Println(err)
			}
			err = w.subClient.Close()
			if err != nil {
				log.Println(err)
			}
		}()
		ch := sub.Channel()
		wg.Done()
		for msg := range ch {
			select {
			case <-w.close:
				return
			default:
			}
			data := msg.Payload
			msgStruct := &MSG{}
			err := msgStruct.UnmarshalBinary([]byte(data))
			if err != nil {
				log.Println(fmt.Printf("Failed to parse message: %s with error: %s\n", data, err.Error()))
			} else {
				isSelf := msgStruct.ID == w.options.LocalID
				if !(w.options.IgnoreSelf && isSelf) {
					log.Printf("receive data: %v", data)
					w.callback(msgStruct)
				}
			}
		}
	}()
	wg.Wait()
}

func (w *Watcher) logRecord(f func() error) error {
	err := f()
	if err != nil {
		log.Println(err)
	}
	return err
}

func (w *Watcher) initConfig(option WatcherOptions, cluster ...bool) error {
	var err error
	if err != nil {
		return err
	}

	if option.SubClient != nil {
		w.subClient = option.SubClient
	} else {
		if len(cluster) > 0 && cluster[0] {
			w.subClient = redis.NewClusterClient(&option.ClusterOptions)
		} else {
			w.subClient = redis.NewClient(&option.Options)
		}
	}

	if option.PubClient != nil {
		w.pubClient = option.PubClient
	} else {
		if len(cluster) > 0 && cluster[0] {
			w.pubClient = redis.NewClusterClient(&option.ClusterOptions)
		} else {
			w.pubClient = redis.NewClient(&option.Options)
		}
	}
	return nil
}

func (w *Watcher) UpdateForSetStr(key string, value string, timeout int64) error {
	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()
		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method:  UpdateForSetStr,
				ID:      w.options.LocalID,
				K:       key,
				V:       value,
				Timeout: timeout,
			}).Err()
	})
}

func (w *Watcher) UpdateForUpdateStr(key string, value string) error {
	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()
		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method: UpdateForUpdateStr,
				ID:     w.options.LocalID,
				K:      key,
				V:      value,
			}).Err()
	})
}

func (w *Watcher) UpdateForSet(key string, value interface{}, timeout int64) error {
	var method UpdateType

	switch t := value.(type) {
	case *model.Session:
		method = UpdateForSetSession
	case *model.QRCode:
		method = UpdateForSetQRCode
	case nil:
		fmt.Printf("nil value: nothing to check?\n")
		return nil
	default:
		fmt.Printf("Unexpected type %T\n", t)
		return nil
	}

	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()

		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method:  method,
				ID:      w.options.LocalID,
				K:       key,
				V:       value,
				Timeout: timeout,
			}).Err()
	})
}

func (w *Watcher) UpdateForUpdate(key string, value interface{}) error {
	var method UpdateType

	switch t := value.(type) {
	case *model.Session:
		method = UpdateForUpdateSession
	case *model.QRCode:
		method = UpdateForUpdateQRCode
	case nil:
		fmt.Printf("nil value: nothing to check?\n")
		return nil
	default:
		fmt.Printf("Unexpected type %T\n", t)
		return nil
	}

	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()
		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method: method,
				ID:     w.options.LocalID,
				K:      key,
				V:      value,
			}).Err()
	})
}

func (w *Watcher) UpdateForDelete(key string) error {
	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()
		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method: UpdateForDelete,
				ID:     w.options.LocalID,
				K:      key,
			}).Err()
	})
}

func (w *Watcher) UpdateForUpdateTimeout(key string, timeout int64) error {
	return w.logRecord(func() error {
		w.l.Lock()
		defer w.l.Unlock()
		return w.pubClient.Publish(
			context.Background(),
			w.options.Channel,
			&MSG{
				Method:  UpdateForUpdateTimeout,
				ID:      w.options.LocalID,
				K:       key,
				Timeout: timeout,
			}).Err()
	})
}
