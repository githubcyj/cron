package worker

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/18 14:46
  @todo
*/

type RedisManager struct {
	RedisPool *redis.Pool
	Conn      redis.Conn
}

var GRedis *RedisManager

func InitRedis() {
	var (
		pool *redis.Pool
		conn redis.Conn
	)
	pool = &redis.Pool{
		Dial: func() (conn redis.Conn, e error) {
			var (
				url string
			)
			url = fmt.Sprintf("%s:%d", GConfig.RedisHost, GConfig.RedisPort)
			c, err := redis.Dial("tcp", url, redis.DialPassword((GConfig.RedisPassword)))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxIdle:     GConfig.RedisMaxIdle,
		MaxActive:   GConfig.RedisMaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
	}
	conn = pool.Get()
	GRedis = &RedisManager{
		RedisPool: pool,
		Conn:      conn,
	}
}
