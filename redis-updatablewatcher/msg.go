package redis_updatablewatcher

import (
	"encoding/json"
	"fmt"
	"github.com/weloe/token-go/model"
)

type MSG struct {
	Method  UpdateType
	ID      string
	K       string
	V       interface{}
	Timeout int64
}

func (m *MSG) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary decodes the struct into a User
func (m *MSG) UnmarshalBinary(data []byte) error {
	type Alias MSG
	alias := &struct {
		*Alias
		V json.RawMessage `json:"V"`
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	switch m.Method {
	case UpdateForSetStr, UpdateForUpdateStr:
		var v string
		if err := json.Unmarshal(alias.V, &v); err != nil {
			return err
		}
		m.V = v
	case UpdateForSetSession, UpdateForUpdateSession:
		var v *model.Session
		if err := json.Unmarshal(alias.V, &v); err != nil {
			return err
		}
		m.V = v
	case UpdateForSetQRCode, UpdateForUpdateQRCode:
		var v *model.QRCode
		if err := json.Unmarshal(alias.V, &v); err != nil {
			return err
		}
		m.V = v
	default:
		return fmt.Errorf("unknown update type: %s", m.Method)
	}

	return nil
}
