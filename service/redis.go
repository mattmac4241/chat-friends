package service

import "gopkg.in/redis.v4"

//REDIS client to be shared throughout service
var REDIS *redis.Client

type redisClient struct{}

//InitRedisClient returns a redis client
func InitRedisClient(address, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return client, err
	}
	return client, nil
}
