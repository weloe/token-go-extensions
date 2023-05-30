package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/weloe/token-go/constant"
	"github.com/weloe/token-go/util"
	"time"
)

/*
	JWT util
*/

/* jwt payload key */
const (
	// LOGIN_TYPE payload key
	LOGIN_TYPE = "loginType"
	LOGIN_ID   = "loginId"
	DEVICE     = "device"
	// EFF expirationTime
	EFF    = "eff"
	RANDOM = "random"
	// NEVER_EXPIRE if expiration time <= -1, return NEVER_EXPIRE
	NEVER_EXPIRE = constant.NeverExpire
	// NOT_VALUE_EXPIRE if expiration time < time.now(), return NOT_VALUE_EXPIRE
	NOT_VALUE_EXPIRE = constant.NotValueExpire
	// EXTRA_DATA extra data key
	EXTRA_DATA = "extraData"
)

// createToken create JWT token and set data
func createToken(loginType string, loginId string, device string, timeout int64, extraData map[string]interface{}, secretKey string) (string, error) {
	// set expiration time
	var expirationTime int64
	if timeout > NEVER_EXPIRE {
		expirationTime = time.Now().UnixMilli() + timeout*1000
	} else {
		expirationTime = timeout
	}

	var claims jwt.MapClaims
	randomString32, err := util.GenerateRandomString32()
	if err != nil {
		return "", err
	}
	// set claims
	if extraData != nil {
		claims = jwt.MapClaims{
			LOGIN_TYPE: loginType,
			LOGIN_ID:   loginId,
			DEVICE:     device,
			EFF:        expirationTime,
			RANDOM:     randomString32,
			EXTRA_DATA: extraData,
		}
	} else {
		claims = jwt.MapClaims{
			LOGIN_TYPE: loginType,
			LOGIN_ID:   loginId,
			DEVICE:     device,
			EFF:        expirationTime,
			RANDOM:     randomString32,
		}
	}

	signature, err := generateToken(claims, secretKey)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func generateToken(claims jwt.MapClaims, secretKey string) (string, error) {

	// sign and get the complete signed token as a string using the secret
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))

	if err != nil {
		return "", err
	}

	return token, nil
}

// parseToken parse token, return JWT payload
func parseToken(token string, loginType string, secretKey string, isCheckTimeout bool) (jwt.MapClaims, error) {

	// secretKey cannot be empty
	if secretKey == "" {
		return nil, errors.New("please configure the JWT secret key")
	}

	// if token is null
	if token == "" {
		return nil, errors.New("JWT string cannot be null")
	}

	// parse
	jwtToken, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		// verify sign alg
		if jwtToken.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("Invalid signing algorithm: " + jwtToken.Method.Alg())
		}
		// return secretKey
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("JWT parsing failed: %v", err)
	}
	payloads, ok := jwtToken.Claims.(jwt.MapClaims)

	// verify token signature
	verifyErr := jwtToken.Claims.Valid()
	if verifyErr != nil {
		return nil, errors.New("Invalid JWT signature: " + token)
	}

	// verify login type
	if !ok || payloads[LOGIN_TYPE] != loginType {
		return nil, errors.New("Invalid JWT login type: " + token)
	}

	// verify Token expiration time
	if isCheckTimeout {
		effFloat, ok := payloads[EFF].(float64)

		if !ok || effFloat < float64(time.Now().UnixMilli()) {
			return nil, errors.New("JWT has expired: " + token)
		}
	}

	return payloads, nil
}

func getId(token string, loginType string, secretKey string) (string, error) {
	payloads, err := parseToken(token, loginType, secretKey, true)
	if err != nil {
		return "", err
	}
	id, ok := payloads[LOGIN_ID].(string)
	if !ok {
		return "", errors.New("Invalid JWT loginId: " + token)
	}
	return id, nil
}

// getTimeout parse and verify loginType return timeout
func getTimeout(token string, loginType string, secretKey string) (int64, error) {
	// parse
	jwtToken, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		// verify sign alg
		if jwtToken.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("Invalid signing algorithm: " + jwtToken.Method.Alg())
		}
		// return secretKey
		return []byte(secretKey), nil
	})
	if err != nil {
		return NOT_VALUE_EXPIRE, errors.New("JWT parsing failed: " + token)
	}
	payloads, ok := jwtToken.Claims.(jwt.MapClaims)

	// verify token signature
	verifyErr := jwtToken.Claims.Valid()
	if verifyErr != nil {
		return NOT_VALUE_EXPIRE, errors.New("Invalid JWT signature: " + token)
	}

	// verify login type
	if !ok || payloads[LOGIN_TYPE] != loginType {
		return NOT_VALUE_EXPIRE, errors.New("Invalid JWT login type: " + token)
	}

	return calTimeout(token, payloads)
}

func calTimeout(token string, payloads jwt.MapClaims) (int64, error) {
	// Convert inputValue to int64
	intValue, ok := payloads[EFF].(float64)
	if !ok {
		return 0, errors.New("Invalid JWT expiration time: " + token)
	}

	if intValue <= float64(NEVER_EXPIRE) {
		return NEVER_EXPIRE, nil
	}

	if intValue < float64(time.Now().UnixMilli()) {
		return NOT_VALUE_EXPIRE, nil
	}

	f := intValue - float64(time.Now().UnixMilli())
	return int64(f / 1000), nil
}
