package buffer

import (
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

func SetupEnqueuer() *work.Enqueuer {
	redisPool := &redis.Pool{
		MaxActive:   10000,
		MaxIdle:     3,
		IdleTimeout: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	return work.NewEnqueuer("query_namespace", redisPool)
}

func SetupCacheConnection() redis.Conn {
	redisPool := &redis.Pool{
		MaxActive:   10000,
		MaxIdle:     3,
		IdleTimeout: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	return redisPool.Get()
}
