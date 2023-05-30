package jwt

import (
	"github.com/golang-jwt/jwt"
	"github.com/weloe/token-go/config"
	"github.com/weloe/token-go/ctx"
	"github.com/weloe/token-go/log"
	"github.com/weloe/token-go/model"
	"github.com/weloe/token-go/persist"
)

var _ IEnforcer = (*StatelessEnforcer)(nil)

type IEnforcer interface {
	// Login login api
	Login(id string, ctx ctx.Context) (string, error)
	LoginByModel(id string, loginModel *model.Login, ctx ctx.Context) (string, error)

	GetLoginId(ctx ctx.Context) (string, error)
	GetClaims(ctx ctx.Context) (jwt.Claims, error)
	GetExtraData(ctx ctx.Context, key string) (interface{}, error)
	GetRequestToken(ctx ctx.Context) string

	GetIdByToken(token string) (string, error)
	GetClaimsByToken(token string) (jwt.Claims, error)
	GetExtraDataByToken(token string, key string) (interface{}, error)

	GetTokenTimeout(token string) (int64, error)

	SetAuth(manager interface{})
	CheckRole(ctx ctx.Context, role string) error
	CheckPermission(ctx ctx.Context, permission string) error

	GetSecretKey() string
	SetSecretKey(secret string)

	SetType(t string)
	GetType() string
	GetAdapter() persist.Adapter
	SetAdapter(adapter persist.Adapter)
	SetWatcher(watcher persist.Watcher)
	GetWatcher() persist.Watcher
	SetLogger(logger log.Logger)
	GetLogger() log.Logger
	EnableLog()
	IsLogEnable() bool
	GetTokenConfig() config.TokenConfig
}
