package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Restaurant struct {
	Name      string
	URL       string
	Price     string
	IsOpenNow bool
	Open      Open
}

type Open struct {
	Day   int
	Start string
	End   string
}

type User struct {
	Name string
}

type Users struct {
	Name []string
}

type DropdownItem struct {
	Name  string
	Value string
}

type Subscribe struct {
	Time  string
	Users []string
}

var redisClient *redis.Client

func main() {
	newRedisClient()
	r := mux.NewRouter()
	r.HandleFunc("/", generateWebsite)
	r.HandleFunc("/subscribe", webSubscribe)
	r.HandleFunc("/index", webIndex)
	r.HandleFunc("/append", webAppend)
	log.Fatal(http.ListenAndServe(":8080", r))

}

func subscribe(restaurant string, user []byte) {
	redisSetList(restaurant, string(user))
}

func getUser() []byte {
	username := "Julian"
	m := User{
		username,
	}
	b, err := json.Marshal(m)
	if err != nil {
	}
	return b
}
