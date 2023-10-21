package redis_updatablewatcher

import (
	"github.com/weloe/token-go/model"
	"testing"
)

func TestMSG_MarshalBinary(t *testing.T) {
	m := &MSG{
		Method: UpdateForUpdateSession,
		ID:     "1",
		K:      "1",
		V: &model.Session{
			Id:            "1",
			Type:          "2",
			LoginType:     "3",
			LoginId:       "4",
			Token:         "5",
			CreateTime:    233,
			DataMap:       nil,
			TokenSignList: nil,
		},
		Timeout: 0,
	}
	bytes, _ := m.MarshalBinary()
	t.Log(string(bytes))
	m2 := &MSG{}
	err := m2.UnmarshalBinary(bytes)
	if err != nil {
		t.Fatalf("Failed to UnmarshalBinary: %v", err)
	}
	t.Log(m2.V.(model.Session))
}

func TestMSG_UnmarshalBinary(t *testing.T) {

}
