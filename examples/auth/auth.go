package auth

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/core/database/redis"
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
		RedisConf redis.RedisConf
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

// NewPeer creates a new peer.
func NewAuthConf(redisNodeList []*RedisNode) *AuthConf {
	return &AuthConf{
		RedisNodeList: redisNodeList,
		MaxCount:      maxCount,
	}

}

func (conf *AuthConf) WithMaxCount(count uint32) *AuthConf {
	conf.MaxCount = count
	return conf
}

func InitAuth(conf *AuthConf) error {
	cHashRing := unsafehash.NewConsistent(conf.MaxCount)
	for _, value := range conf.RedisNodeList {
		if err := value.RedisConf.Validate(); err != nil {
			return err
		}
		redis := redis.NewRedis(value.RedisConf.Host, value.RedisConf.Type, value.RedisConf.Pass)
		cHashRing.Add(unsafehash.NewNode(value.RedisConf.Host, value.Position, redis))
	}
	AuthManager.cHashRing = cHashRing
	return nil
}

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
	value, err := node.Extra.(*redis.Redis).Get(prefix + arr[0])
	if value != token {
		return "", TokenInValid
	}
	if err != nil {
		return "", TokenNotFound
	}
	return arr[0], nil
}

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
	err = node.Extra.(*redis.Redis).Set(prefix+id, token)
	if err != nil {
		return "", err
	}
	return token, nil
}
