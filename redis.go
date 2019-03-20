package main

import (
	"github.com/go-redis/redis"
	"github.com/tidwall/sjson"
)

func newRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func redisSet(key string, value string) {
	err := redisClient.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}

}

func redisSetList(key string, value string) {
	err := redisClient.RPush(key, value).Err()
	if err != nil {
		panic(err)
	}
	err = redisClient.SortStore(key, key, &redis.Sort{By: "Fa*"}).Err()
	if err != nil {
		panic(err)
	}

}

func redisGet(key string) string {
	values, _ := redisClient.Get(key).Result()
	return values
}

func redisAppend(key string, index int, user string) {
	bla, _ := redisClient.LRange(key, int64(index), int64(index)).Result()
	value, _ := sjson.Set(bla[0], "Users.-1", user)
	err := redisClient.LSet(key, int64(index), value).Err()
	if err != nil {
		panic(err)
	}
}

func redisGetKeys(prefix string) []string {
	keys, _ := redisClient.Keys(prefix).Result()
	return keys
}

func redisListLength(key string) int64 {
	length, _ := redisClient.LLen(key).Result()
	return length
}

func redisGetList(key string, start int64, stop int64) []string {
	values, err := redisClient.LRange(key, start, stop).Result()
	if err != nil {
		panic(err)
	}
	return values
}
