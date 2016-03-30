/*
Package redis provides the library to communicate to redis
*/
package redis

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var IDLE_TIMEOUT = 240 * time.Second

// Connect initializes the redis connection pool
func Connect(host string, maxactive, maxidle int) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: IDLE_TIMEOUT,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	c := pool.Get()
	if _, err := c.Do("PING"); err != nil {
		log.Fatal("unable to connect to redis: ", host, " err: ", err)
	}
	return &pool
}

//AuthConnect connects and authenticated with the password
func AuthConnect(host string, password string, maxactive, maxidle int) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: IDLE_TIMEOUT,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	c := pool.Get()
	if _, err := c.Do("PING"); err != nil {
		log.Fatal("unable to connect to redis: ", host, " err: ", err)
	}
	return &pool
}
