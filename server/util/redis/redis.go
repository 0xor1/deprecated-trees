package redis

import (
	"errors"
	"github.com/0xor1/iredis"
	"github.com/garyburd/redigo/redis"
	"time"
)

func CreatePool(address string) iredis.Pool {
	return &redis.Pool{
		MaxIdle:     300,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, redis.DialDatabase(0), redis.DialConnectTimeout(500*time.Millisecond), redis.DialReadTimeout(500*time.Millisecond), redis.DialWriteTimeout(500*time.Millisecond))
		},
		TestOnBorrow: func(c redis.Conn, ti time.Time) error {
			if time.Since(ti) < time.Minute {
				return nil
			}
			return errors.New("Redis connection timed out")
		},
	}
}
