# token-go-extensions
token-go extensions support adapter, watcher ...

## redis-adapter
`go get github.com/weloe/token-go-extensions/redis-adapter`

use Redis to store data

### Usage
```go
import (
    tokenGo "github.com/weloe/token-go"
    redisadapter "github.com/weloe/token-go-extensions/redis-adapter"
)

var (
    TokenEnforcer *tokenGo.Enforcer
)

func CreateTokenEnforcer() {
    var err error
    adapter, err := redisadapter.NewAdapter("ip:port", "password", dbNum)
    if err != nil {
        log.Fatalf("NewRedisAdapter() failed: %v", err)
    }
    TokenEnforcer, err = tokenGo.NewEnforcer(adapter)
    if err != nil {
        log.Fatalf("NewEnforcer() failed: %v", err)
    }
}
```
## jwt
`go get github.com/weloe/token-go-extensions/jwt`

StatelessEnforcer used `github.com/golang-jwt/jwt` to generate and parse token

### Usage
```go
import (
    "github.com/weloe/token-go-extensions/jwt"
)

func main() {
    enforcer, err := jwt.NewEnforcer()
	if err != nil {
		log.Printf("NewEnforcer() failed: %v", err)
	}
	token, err := enforcer.Login("1", nil)
	if err != nil {
		log.Printf("Login() failed: %v", err)
	} else {
		log.Printf("login success, token = %v", token)
	}
}
```

## redis-updatablewatcher
`go get github.com/weloe/token-go-extensions/redis-updatablewatcher`

use redis publish/subscribe implement UpdatableWatcher

