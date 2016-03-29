/*
Package redis provides the library to communicate to redis
*/
package redis

import (
	"github.com/garyburd/redigo/redis"
)

var (
	pool *redis.Pool
)

// Connect initializes the redis connection pool
func Connect(host string, maxcon int) {
	pool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", host)

		if err != nil {
			return nil, err
		}

		return c, err
	}, maxcon)
}

//Get fetches a connection from the pool
func Get() redis.Conn {
	return pool.Get()
}

//Close closes the connection pool
func Close() {
	pool.Close()
}
