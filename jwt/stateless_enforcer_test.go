package jwt

import (
	"testing"
)

func newTestEnforcer(t *testing.T) *StatelessEnforcer {
	enforcer, err := NewEnforcer()
	if err != nil {
		t.Errorf("NewEnforcer() failed: %v", err)
	}
	t.Log(enforcer)
	return enforcer
}

func TestStatelessEnforcer_Login(t *testing.T) {
	enforcer := newTestEnforcer(t)

	enforcer.SetSecretKey("123")

	token, err := enforcer.Login("1", nil)
	if err != nil {
		t.Errorf("Login() failed: %v", err)
	}
	t.Logf("generate token: %v", token)

	data, err := enforcer.GetExtraDataByToken(token, "123")
	if err != nil {
		t.Errorf("GetExtraDataByToken() failed: %v", err)
	}
	if data != nil {
		t.Errorf("Unexpected data: %v", data)
	}

}
