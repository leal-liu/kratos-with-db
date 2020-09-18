package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClientWrapper
type RedisClientWrapper struct {
	client *redis.Client
}

// set execute redis SET command cache data
func (object *RedisClientWrapper) set(cacheKey string,
	cache []byte,
	expiration, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = object.client.Set(ctx, cacheKey, cache, expiration).Result()
	return
}

// setObject set object
func (object *RedisClientWrapper) setObject(cacheKey string,
	instance interface{},
	expiration, timeout time.Duration) (err error) {
	var cache []byte
	if cache, err = json.Marshal(instance); nil != err {
		panic(err)
	}
	err = object.set(cacheKey, cache, expiration, timeout)
	return
}

// get execute redis GET command get cache data
func (object *RedisClientWrapper) get(cacheKey string,
	timeout time.Duration) (cache string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cache, err = object.client.Get(ctx, cacheKey).Result()
	if redis.Nil == err {
		err = nil
	}
	return
}

// getObject get object
func (object *RedisClientWrapper) getObject(cacheKey string,
	timeout time.Duration, instance interface{}) (err error, ok bool) {
	var cache string
	if cache, err = object.get(cacheKey, timeout); nil != err {
		return
	}
	if 0 >= len(cache) {
		return
	}
	if err = json.Unmarshal([]byte(cache), instance); nil != err {
		panic(err)
	}
	ok = true
	return
}

// hSet execute redis HSET command cache data
func (object *RedisClientWrapper) hSet(cacheKey, fieldKey string,
	cache []byte,
	timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = object.client.HSet(ctx, cacheKey, fieldKey, cache).Result()
	return
}

// hSetObject set object
func (object *RedisClientWrapper) hSetObject(cacheKey, fieldKey string,
	instance interface{},
	timeout time.Duration) (err error) {
	var cache []byte
	if cache, err = json.Marshal(instance); nil != err {
		panic(err)
	}
	err = object.hSet(cacheKey, fieldKey, cache, timeout)
	return
}

// hGet execute redis HGET command get cache data
func (object *RedisClientWrapper) hGet(cacheKey, fieldKey string,
	timeout time.Duration) (cache string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cache, err = object.client.HGet(ctx, cacheKey, fieldKey).Result()
	if redis.Nil == err {
		err = nil
	}
	return
}

// hGetObject get object
func (object *RedisClientWrapper) hGetObject(cacheKey, fieldKey string,
	instance interface{},
	timeout time.Duration) (err error, ok bool) {
	var cache string
	if cache, err = object.hGet(cacheKey, fieldKey, timeout); nil != err {
		return
	}
	if 0 >= len(cache) {
		return
	}
	if err = json.Unmarshal([]byte(cache), instance); nil != err {
		panic(err)
	}
	ok = true
	return
}

func (object *RedisClientWrapper) del(cacheKey string, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = object.client.Del(ctx, cacheKey).Result()

	if err != nil && err == redis.Nil {
		return nil
	}

	return
}

func (object *RedisClientWrapper) DelDelegationRewardModel(cacheKey string, timeout time.Duration) (err error) {
	err = object.del(cacheKey, timeout)
	return
}

func (object *RedisClientWrapper) DelAccountCoinsModel(cacheKey string, timeout time.Duration) (err error) {
	err = object.del(cacheKey, timeout)
	return
}

// NewRedisClientWrapper factory method
func NewRedisClientWrapper(client *redis.Client) *RedisClientWrapper {
	return &RedisClientWrapper{client: client}
}
