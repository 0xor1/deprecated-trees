package iredis

import(
	"github.com/gomodule/redigo/redis"
)

type Pool interface {
	Get() Conn
}

type Conn interface{
	redis.Conn
}

func NewPool(p *redis.Pool) Pool {
	return &pool{p}
}

type pool struct{
	pool *redis.Pool
}

func (p *pool) Get() Conn {
	return p.pool.Get()
}