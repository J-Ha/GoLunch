package main

import (
	"fmt"

	"github.com/go-redis/redis"
)

func AAAAAAAmain() {
	ExampleNewClient("bla", "foo")
}

func ExampleNewClient(key string, value string) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := client.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}

	bla := client.Get(key)
	fmt.Println(bla)

	err2 := client.Append(key, ",lalalalala")
	if err != nil {
		panic(err2)
	}

	bla = client.Get(key)
	fmt.Println(bla)
}
