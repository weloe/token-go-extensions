package redis_updatablewatcher

import (
	"github.com/go-redis/redis/v8"
	tokenGo "github.com/weloe/token-go"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
	"testing"
	"time"
)

func initWatcherWithOptions(t *testing.T, wo WatcherOptions, cluster ...bool) (*tokenGo.Enforcer, *Watcher) {
	var (
		w   persist.UpdatableWatcher
		err error
	)
	if len(cluster) > 0 && cluster[0] {
		w, err = NewWatcherWithCluster("127.0.0.1:6379,127.0.0.1:6379,127.0.0.1:6379", wo)
	} else {
		w, err = NewWatcher("127.0.0.1:6379", wo)
	}
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	e, err := tokenGo.NewEnforcer(tokenGo.NewDefaultAdapter())
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}
	e.SetUpdatableWatcher(w)
	_ = w.(*Watcher).SetUpdateCallback(DefaultUpdateCallback(e))
	return e, w.(*Watcher)
}

func initWatcher(t *testing.T, cluster ...bool) (*tokenGo.Enforcer, *Watcher) {
	return initWatcherWithOptions(t, WatcherOptions{Options: redis.Options{
		Password: "",
		DB:       2,
	}}, cluster...)
}

func TestWatcher_UpdateForSetSession(t *testing.T) {
	enforcer, _ := initWatcher(t, false)
	session := &model.Session{
		Id:        "1",
		Type:      "t1",
		LoginType: "l1",
		LoginId:   "testLoginId1",
		Token:     "t1",
	}
	err := enforcer.SetSession("1", session, -1)
	if err != nil {
		t.Fatalf("Failed to SetSession: %v", err)
	}
	time.Sleep(time.Second)
	s := enforcer.GetSession("1")
	if s != nil {
		t.Log(s.Json())
	}
	session.Token = "t2"
	err = enforcer.UpdateSession("1", session)
	if err != nil {
		t.Fatalf("Failed to UpdateSession: %v", err)
	}
	time.Sleep(time.Second)
	s1 := enforcer.GetSession("1")
	if s1 != nil {
		t.Log(s1.Json())
	}
}

func TestMsg(t *testing.T) {
	m := &MSG{
		Method:  "m1",
		ID:      "id1",
		K:       "k1",
		V:       "nil",
		Timeout: -1,
	}
	bytes, _ := m.MarshalBinary()
	t.Log(string(bytes))
}
