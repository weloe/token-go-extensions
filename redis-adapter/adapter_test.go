package redis_adapter

import (
	tokengo "github.com/weloe/token-go"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
	"testing"
	"time"
)

func newTestRedisAdapter(t *testing.T) persist.Adapter {
	addr := "127.0.0.1:6379"
	username := ""
	pwd := "pwd"
	db := 1
	adapter, err := NewAdapter(addr, username, pwd, db)
	if err != nil {
		t.Fatalf("NewAdapter() failed: %v", err)
	}
	return adapter
}

func TestDefaultAdapter_StrOperation(t *testing.T) {
	adapter := newTestRedisAdapter(t)

	if v := adapter.GetStrTimeout("unExist"); v != 0 {
		t.Fatalf("GetStrTimeout() failed: timeout is %v,want 0 ", v)
	}

	if err := adapter.SetStr("k1", "v1", 0); err != nil {
		t.Fatalf("SetStr() failed: set timeout = 0")
	}

	if err := adapter.SetStr("k2", "v2", -1); err != nil {
		t.Fatalf("SetStr() failed: can't set data")
	}

	if v := adapter.GetStr("k2"); v != "v2" {
		t.Fatalf("GetStr() failed: value is %s, want 'v2' ", v)
	}

	if v := adapter.GetStrTimeout("k2"); v != 0 {
		t.Fatalf("GetStrTimeout() failed: timeout is %v,want 0 ", v)
	}

	if err := adapter.SetStr("k1", "v1", 1); err != nil {
		t.Fatalf("SetStr() failed: can't set data")
	}
	time.Sleep(2 * time.Second)
	if v := adapter.Get("k1"); v != nil {
		t.Fatalf("getExpireAndDelete() faliled: get expired value")
	}

	err1 := adapter.SetStr("k", "v", -1)
	if err1 != nil {
		t.Fatalf("SetStr() failed: %v", err1)
	}

	if err := adapter.UpdateStrTimeout("k", 9); err != nil {
		t.Fatalf("UpdateStrTimeout() failed: %v", err)
	}

	timeout := adapter.GetStrTimeout("k")
	t.Logf("get timeout = %v", timeout)

	getRes := adapter.GetStr("k")
	if getRes != "v" {
		t.Fatalf("GetStr() failed: %v", getRes)
	}

	err3 := adapter.UpdateStr("k", "L")
	if err3 != nil {
		t.Fatalf("UpdateStr() failed: %v", err3)
	}

	getRes = adapter.GetStr("k")
	if getRes != "L" {
		t.Fatalf("GetStr() failed: GetStr() =  %v want 'L' ", getRes)
	}

	err4 := adapter.DeleteStr("k")
	if err4 != nil {
		t.Fatalf("DeleteStr() failed: %v", err4)
	}

	getRes = adapter.GetStr("k")
	if getRes != "" {
		t.Fatalf("GetStr() failed: %v", getRes)
	}
}

func TestDefaultAdapter_InterfaceOperation(t *testing.T) {
	defaultAdapter := newTestRedisAdapter(t)

	if v := defaultAdapter.GetTimeout("unExist"); v != 0 {
		t.Fatalf("GetTimeout() failed: timeout is %v,want 0 ", v)
	}

	if err := defaultAdapter.Set("k1", "v1", 0); err != nil {
		t.Fatalf("Set() failed: set timeout = 0")
	}

	if err := defaultAdapter.Set("k2", "v2", -1); err != nil {
		t.Fatalf("Set() failed: can't set data")
	}

	if v := defaultAdapter.Get("k2"); v != nil {
		t.Fatalf("Get() failed: value is %s, want 'v2' ", v)
	}

	if v := defaultAdapter.GetTimeout("k2"); v != 0 {
		t.Fatalf("GetTimeout() failed: timeout is %v,want 0 ", v)
	}

	if err := defaultAdapter.Set("k1", "v1", 1); err != nil {
		t.Fatalf("Set() failed: can't set data")
	}
	time.Sleep(2 * time.Second)
	if v := defaultAdapter.Get("k1"); v != nil {
		t.Fatalf("Get() faliled: get expired value")
	}

	err1 := defaultAdapter.Set("k", "v", -1)
	if err1 != nil {
		t.Fatalf("Set() failed: %v", err1)
	}

	if err := defaultAdapter.UpdateTimeout("k", 100); err != nil {
		t.Fatalf("UpdateTimeout() failed: %v", err)
	}

	timeout := defaultAdapter.GetTimeout("k")
	t.Logf("get timeout = %v", timeout)

	getRes := defaultAdapter.Get("k")
	if getRes != nil {
		t.Fatalf("GetGetStr() failed: %v", getRes)
	}

	err3 := defaultAdapter.Update("k", "L")
	if err3 != nil {
		t.Fatalf("Update() failed: %v", err3)
	}

	getRes = defaultAdapter.Get("k")
	if getRes != nil {
		t.Fatalf("Get() failed: GetStr() =  %v want 'L' ", getRes)
	}

	err4 := defaultAdapter.Delete("k")
	if err4 != nil {
		t.Fatalf("Delete() failed: %v", err4)
	}

	getRes = defaultAdapter.Get("k")
	if getRes != nil {
		t.Fatalf("Get() failed: %v", getRes)
	}
}

func TestDefaultAdapter_DeleteBatchFilteredValue(t *testing.T) {
	adapter := newTestRedisAdapter(t)
	if err := adapter.SetStr("k_1", "v", -1); err != nil {
		t.Errorf("SetStr() failed: %v", err)
	}
	if err := adapter.SetStr("k_2", "v", -1); err != nil {
		t.Errorf("SetStr() failed: %v", err)
	}
	if err := adapter.SetStr("k_3", "v", -1); err != nil {
		t.Errorf("SetStr() failed: %v", err)
	}
	err := adapter.(persist.BatchAdapter).DeleteBatchFilteredKey("k_")
	if err != nil {
		t.Errorf("DeleteBatchFilteredKey() failed: %v", err)
	}
	str := adapter.GetStr("k_2")
	if str != "" {
		t.Errorf("DeleteBatchFilteredKey() failed")
	}
}

func TestEnforcer(t *testing.T) {
	enforcer, err := tokengo.NewEnforcer(newTestRedisAdapter(t))
	if err != nil {
		t.Errorf("NewEnforcer() failed: %v", err)
	}
	session := model.DefaultSession("1")
	session.AddTokenSign(&model.TokenSign{
		Value:  "token-value-1",
		Device: "web",
	})
	session.AddTokenSign(&model.TokenSign{
		Value:  "token-value-2",
		Device: "mobile",
	})
	err = enforcer.SetSession("1", session, 233)
	if err != nil {
		t.Errorf("enforcer.SetSession() failed: %v", err)
	}
	getSession := enforcer.GetSession("1")
	t.Log(getSession.TokenSignList)
	if getSession.TokenSignSize() != 2 {
		t.Errorf("unexpected tokenSignSize = %v ", getSession.TokenSignSize())
	}

}
