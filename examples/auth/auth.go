package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/crypto/aes"
)

const maxCount = 1000

type (
	AuthConf struct {
		Auth          bool
		RedisNodeList []*RedisNode
		MaxCount      uint32
	}
	RedisNode struct {
		RedisConf *redis.Options
		Position  uint32
	}
	Auth struct {
		Auth      bool
		cHashRing *unsafehash.Consistent
	}
)

var (
	AuthManager   = new(Auth)
	TokenNotFound = fmt.Errorf("token not found")
	TokenInValid  = fmt.Errorf("token is invalid")
	prefix        = "token#"
)

// NewAuthConf
func NewAuthConf(redisNodeList []*RedisNode) *AuthConf {
	return &AuthConf{
		RedisNodeList: redisNodeList,
		MaxCount:      maxCount,
	}

}

// WithMaxCount
func (conf *AuthConf) WithMaxCount(count uint32) *AuthConf {
	conf.MaxCount = count
	return conf
}

// InitAuth
func InitAuth(conf *AuthConf) error {
	cHashRing := unsafehash.NewConsistent(conf.MaxCount)
	for _, value := range conf.RedisNodeList {
		redis := redis.NewClient(value.RedisConf)
		cHashRing.Add(unsafehash.NewNode(value.RedisConf.Addr, value.Position, redis))
	}
	AuthManager.cHashRing = cHashRing
	return nil
}

// Validate
func (auth *Auth) Validate(token string) (string, error) {
	if token == "" {
		return "", TokenNotFound
	}
	tokenRaw, err := aes.NewAes([]byte("lgqgg56oi9a9tefl")).Decrypt(token)
	arr := strings.Split(tokenRaw, ":")
	if len(arr) < 2 {
		return "", TokenNotFound
	}
	if err != nil {
		return "", TokenNotFound
	}
	node := auth.cHashRing.Get(arr[0])
	value, err := node.Extra.(*redis.Client).Get(context.Background(), prefix+arr[0]).Result()
	if value != token {
		return "", TokenInValid
	}
	if err != nil {
		return "", TokenNotFound
	}
	return arr[0], nil
}

// Add
func (auth *Auth) Add(id string) (string, error) {
	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)
	tokenRaw := id + ":" + nowStr
	tokenBytes, err := aes.NewAes([]byte("lgqgg56oi9a9tefl")).Encrypt(tokenRaw)
	token := string(tokenBytes)
	if err != nil {
		return "", err
	}
	node := auth.cHashRing.Get(id)
	err = node.Extra.(*redis.Client).Set(context.Background(), prefix+id, token, 0).Err()
	if err != nil {
		return "", err
	}
	return token, nil
}
