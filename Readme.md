# token-go-extensions
token-go extensions support adapter, watcher ...

## redis-adapter
`go get github.com/weloe/token-go-extensions/redis-adapter`

use Redis to store data

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
