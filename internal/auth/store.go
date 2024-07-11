package auth

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisClient interface {
	Get(key string) (*AccessToken, error)
	Set(key, value *AccessToken) error
	Delete(key string) error
	Expire(key string, seconds int) (int, error)
}

type PoolRedisClient struct {
	pool *redis.Pool
}

func NewPoolRedisClient(address string) (*PoolRedisClient, error) {
	pool := &redis.Pool{
		MaxIdle:   5,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}

			if _, err := conn.Do("SELECT", 1); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
	}

	conn := pool.Get()
	defer conn.Close()
	if _, err := conn.Do("PING"); err != nil {
		return nil, err
	}

	return &PoolRedisClient{pool}, nil
}

func (p *PoolRedisClient) Close() error {
	return p.pool.Close()
}

func (p *PoolRedisClient) Get(key string) (string, error) {
	conn := p.pool.Get()
	defer conn.Close()
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	return value, nil
}

// func (p *PoolRedisClient) Set(key string, value []byte, aexp time.Duration) error {
func (p *PoolRedisClient) Set(key, value string, aexp time.Duration) error {
	conn := p.pool.Get()
	defer conn.Close()

	if aexp == 0 {
		_, err := conn.Do("SET", key, value)
		return err
	}

	aexpSeconds := int(aexp.Seconds())
	_, err := conn.Do("SET", key, value, "EX", aexpSeconds)
	return err
}

func (p *PoolRedisClient) HMSet(key string, values map[string]interface{}, aexp time.Duration) error {
	conn := p.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for k, v := range values {
		args = args.Add(k, v)
	}

	if aexp > 0 {
		aexpSeconds := int(aexp.Seconds())
		if _, err := conn.Do("EXPIRE", key, aexpSeconds); err != nil {
			return err
		}
	}

	if _, err := conn.Do("HMSET", args...); err != nil {
		return err
	}
	return nil
}

func (p *PoolRedisClient) HGetAll(key string) (map[string]string, error) {
	conn := p.pool.Get()
	defer conn.Close()
	values, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (p *PoolRedisClient) Expire(key string, seconds int) (int, error) {
	conn := p.pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("EXPIRE", key, seconds))
}

func (p *PoolRedisClient) Del(key string) error {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}
