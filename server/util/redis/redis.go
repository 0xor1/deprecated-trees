package redis

import (
	"errors"
	"github.com/0xor1/iredis"
	"github.com/garyburd/redigo/redis"
	"time"
)

func CreatePool(address string, log func(error)) iredis.Pool {
	return &redis.Pool{
		MaxIdle:     300,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			conn, e := redis.Dial("tcp", address, redis.DialDatabase(0), redis.DialConnectTimeout(1*time.Second), redis.DialReadTimeout(2*time.Second), redis.DialWriteTimeout(2*time.Second))
			// Log any Redis connection error on stdout
			if e != nil {
				log(e)
			}

			return conn, e
		},
		TestOnBorrow: func(c redis.Conn, ti time.Time) error {
			if time.Since(ti) < time.Minute {
				return nil
			}
			return errors.New("Redis connection timed out")
		},
	}
}
