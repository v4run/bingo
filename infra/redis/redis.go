/*
Package redis provides the library to communicate to redis
*/
package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var IDLE_TIMEOUT = 240 * time.Second

// Connect initializes the redis connection pool
func Connect(host string, maxactive, maxidle int) (*redis.Pool, error) {
	pool := &redis.Pool{
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
	return check(host, pool)
}

//AuthConnect connects and authenticated with the password
func AuthConnect(host string, password string, maxactive, maxidle int) (*redis.Pool, error) {
	pool := &redis.Pool{
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
	return check(host, pool)
}

func check(host string, pool *redis.Pool) (*redis.Pool, error) {
	c := pool.Get()
	if _, err := c.Do("PING"); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %s err: %s", host, err)
	}
	return pool, nil
}
