package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	tokenGo "github.com/weloe/token-go"
	"github.com/weloe/token-go/config"
	"github.com/weloe/token-go/ctx"
	"github.com/weloe/token-go/log"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
)

// StatelessEnforcer use Jwt implement
type StatelessEnforcer struct {
	e *tokenGo.Enforcer
}

func (s *StatelessEnforcer) SetAuth(manager interface{}) {
	s.e.SetAuth(manager)
}

func (s *StatelessEnforcer) CheckRole(ctx ctx.Context, role string) error {
	return s.e.CheckRole(ctx, role)
}

func (s *StatelessEnforcer) CheckPermission(ctx ctx.Context, permission string) error {
	return s.e.CheckPermission(ctx, permission)
}

func (s *StatelessEnforcer) GetAdapter() persist.Adapter {
	return s.e.GetAdapter()
}

func (s *StatelessEnforcer) SetAdapter(adapter persist.Adapter) {
	s.e.SetAdapter(adapter)
}

func (s *StatelessEnforcer) EnableLog() {
	s.e.EnableLog()
}

func (s *StatelessEnforcer) IsLogEnable() bool {
	return s.e.IsLogEnable()
}

func (s *StatelessEnforcer) GetTokenConfig() config.TokenConfig {
	return s.e.GetTokenConfig()
}

// NewEnforcer new jwt enforcer, parameter need TokenConfig or string
func NewEnforcer(args ...interface{}) (*StatelessEnforcer, error) {
	var e *tokenGo.Enforcer
	var err error
	if len(args) > 0 {
		e, err = tokenGo.NewEnforcer(&persist.EmptyAdapter{}, args[0])
	} else {
		e, err = tokenGo.NewEnforcer(&persist.EmptyAdapter{})
	}
	if err != nil {
		return nil, err
	}
	return &StatelessEnforcer{e}, nil
}

func (s *StatelessEnforcer) SetType(t string) {
	s.e.SetType(t)
}

func (s *StatelessEnforcer) GetType() string {
	return s.e.GetType()
}

func (s *StatelessEnforcer) SetLogger(logger log.Logger) {
	s.e.SetLogger(logger)
}

func (s *StatelessEnforcer) GetLogger() log.Logger {
	return s.e.GetLogger()
}

func (s *StatelessEnforcer) SetWatcher(watcher persist.Watcher) {
	s.e.SetWatcher(watcher)
}

func (s *StatelessEnforcer) GetWatcher() persist.Watcher {
	return s.e.GetWatcher()
}

func (s *StatelessEnforcer) SetSecretKey(secret string) {
	s.e.SetJwtSecretKey(secret)
}

func (s *StatelessEnforcer) GetSecretKey() string {
	return s.e.GetTokenConfig().JwtSecretKey
}

// Login loginById and loginModel, return tokenValue and error
// ctx.Context can be nil
func (s *StatelessEnforcer) Login(id string, ctx ctx.Context) (string, error) {
	return s.LoginByModel(id, model.DefaultLoginModel(), ctx)
}

// LoginByModel login by id and loginModel, return tokenValue and error
// ctx.Context can be nil
func (s *StatelessEnforcer) LoginByModel(id string, loginModel *model.Login, ctx ctx.Context) (string, error) {
	if loginModel == nil {
		return "", errors.New("arg loginModel can not be nil")
	}
	token, err := createToken(s.e.GetType(), id, loginModel.Device, loginModel.Timeout, loginModel.JwtData, s.GetSecretKey())
	if err != nil {
		return "", err
	}

	err = s.e.ResponseToken(token, loginModel, ctx)
	if err != nil {
		return "", err
	}

	// called watcher
	m := &model.Login{
		Device:          loginModel.Device,
		IsLastingCookie: loginModel.IsLastingCookie,
		Timeout:         loginModel.Timeout,
		JwtData:         loginModel.JwtData,
		Token:           token,
		IsWriteHeader:   loginModel.IsWriteHeader,
	}

	// called logger
	s.e.GetLogger().Login(s.e.GetType(), id, token, m)

	// called watcher
	if s.e.GetWatcher() != nil {
		s.e.GetWatcher().Login(s.e.GetType(), id, token, m)
	}

	return token, nil
}

// GetRequestToken get token from request
func (s *StatelessEnforcer) GetRequestToken(ctx ctx.Context) string {
	return s.e.GetRequestToken(ctx)
}

// GetClaims get claims by web context
func (s *StatelessEnforcer) GetClaims(ctx ctx.Context) (jwt.Claims, error) {
	token := s.GetRequestToken(ctx)
	if token == "" {
		return nil, errors.New("token is nil")
	}
	return s.GetClaimsByToken(token)
}

// GetExtraData get extra data by web context
func (s *StatelessEnforcer) GetExtraData(ctx ctx.Context, key string) (interface{}, error) {
	token := s.GetRequestToken(ctx)
	if token == "" {
		return nil, errors.New("token is nil")
	}
	return s.GetExtraDataByToken(token, key)
}

// GetClaimsByToken get token claims
func (s *StatelessEnforcer) GetClaimsByToken(token string) (jwt.Claims, error) {
	return parseToken(token, s.GetType(), s.GetSecretKey(), true)
}

// GetExtraDataByToken parse extraData map
func (s *StatelessEnforcer) GetExtraDataByToken(token string, key string) (interface{}, error) {
	mapClaims, err := parseToken(token, s.GetType(), s.GetSecretKey(), true)
	if err != nil {
		return nil, err
	}
	extraData := mapClaims[EXTRA_DATA]
	if extraData == nil {
		return nil, nil
	}
	extraMap, ok := extraData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid extraData: %v", extraData)
	}
	return extraMap[key], nil
}

func (s *StatelessEnforcer) GetLoginId(ctx ctx.Context) (string, error) {
	token := s.GetRequestToken(ctx)
	return s.GetIdByToken(token)
}

// GetIdByToken parse token and get id
func (s *StatelessEnforcer) GetIdByToken(token string) (string, error) {
	return getId(token, s.GetType(), s.GetSecretKey())
}

// GetTokenTimeout parse and get token timeout
func (s *StatelessEnforcer) GetTokenTimeout(token string) (int64, error) {
	timeout, err := getTimeout(token, s.GetType(), s.GetSecretKey())
	if err != nil {
		return 0, err
	}
	return timeout, nil
}

// GetTokenTimeoutByCtx similar with GetTokenTimeout
func (s *StatelessEnforcer) GetTokenTimeoutByCtx(ctx ctx.Context) (int64, error) {
	token := s.e.GetRequestToken(ctx)
	return s.GetTokenTimeout(token)
}
