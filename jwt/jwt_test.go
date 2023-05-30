package jwt

import "testing"

func TestJwt(t *testing.T) {
	m := make(map[string]interface{})
	m["1"] = "v1"
	m["2"] = "v2"
	m["3"] = "v3"
	correctKey := "proper"
	errorKey := "error"
	token, err := createToken("user", "1", "device", 22, m, correctKey)
	if err != nil {
		t.Errorf("createToken() failed: %v", err)
	}
	t.Logf("create token = %v", token)

	_, err = getId(token, "device", errorKey)
	if err == nil {
		t.Errorf("GetIdByToken() failed: %v", err)
	}

	id, err := getId(token, "user", correctKey)
	if err != nil {
		t.Errorf("GetIdByToken() failed: %v", err)
	}
	t.Logf("get id = %v", id)
	if id != "1" {
		t.Errorf("GetIdByToken() failed: unexpected id %v", id)
	}

	timeout, err := getTimeout(token, "user", correctKey)
	if err != nil {
		t.Errorf("GetTokenTimeout() failed: %v", err)
	}
	t.Logf("timeout = %v", timeout)
}
