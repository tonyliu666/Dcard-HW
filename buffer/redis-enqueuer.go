package buffer

import (
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

func SetupEnqueuer() (*work.Enqueuer) {
	redisPool := &redis.Pool{
		MaxActive: 7000,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	return work.NewEnqueuer("query_namespace", redisPool)
}

func SetupCacheConnection() (redis.Conn) {
	redisPool := &redis.Pool{
		MaxActive: 7000,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	return redisPool.Get()
}