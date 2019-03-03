package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

var query struct {
	Search struct {
		Total    graphql.Int
		Business []struct {
			Name   graphql.String
			Rating graphql.Float
			Price  graphql.String
			Url    graphql.String
			Hours  []struct {
				Is_open_now graphql.Boolean
				Open        []struct {
					Day   graphql.Int
					Start graphql.String
					End   graphql.String
				}
			}
		}
	} `graphql:"search(location: $zip, radius: 1500, limit: $limit, offset: $offset)"`
}

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
	http.HandleFunc("/", generateWebsite)
	http.HandleFunc("/subscribe", webSubscribe)
	http.HandleFunc("/index", webIndex)
	http.HandleFunc("/append", webAppend)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func yelpSearch(limit int, offset int) {
	variables := map[string]interface{}{
		"zip":    graphql.String(os.Getenv("GOLUNCH_ZIP")),
		"limit":  graphql.Int(limit),
		"offset": graphql.Int(offset),
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := graphql.NewClient("https://api.yelp.com/v3/graphql", httpClient)

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
	}
}

func indexRestaurants() {
	offset := 0
	limit := 50
	yelpSearch(limit, offset)
	num := int(query.Search.Total) / limit
	fmt.Println(query.Search.Total)
	for i := 0; i <= num; i++ {
		yelpSearch(limit, offset)
		for _, business := range query.Search.Business {
			open := Open{}
			opennow := true
			m := Restaurant{
				string(business.Name),
				string(business.Url),
				string(business.Price),
				opennow,
				open,
			}
			b, err := json.Marshal(m)
			if err != nil {
				// Handle error.
			}
			redisSet("r/"+string(business.Name), string(b))
		}
		offset = offset + limit
	}
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

func subscribe(restaurant string, user []byte) {
	redisSetList("s/"+restaurant, string(user))
}

func generateWebsite(w http.ResponseWriter, r *http.Request) {
	names, _ := redisGetKeys("r/*").Result()
	html, _ := ioutil.ReadFile("template.html")
	cont, _ := ioutil.ReadFile("content.html")

	var HtmlRest = make(map[string]interface{})
	for num, rest := range names {
		values, _ := redisGet(rest).Result()
		HtmlRest[strings.Replace(names[num], "r/", "", -1)] = values
	}

	subs, _ := redisGetKeys("s/*").Result()
	var HtmlSubs = make(map[string]interface{})
	for _, rests := range subs {
		subRest, _ := redisGet(strings.Replace(strings.TrimRight(strings.SplitAfter(rests, "-")[0], "-"), "s/", "r/", -1)).Result()
		subss, _ := redisClient.LRange(rests, 0, -1).Result()
		HtmlSubs[subRest] = subss
	}
	dropdownTemplate, err := template.New("dropdownexample").Parse(string(html))
	if err != nil {
		panic(err)
	}
	dropdownTemplate.Execute(w, HtmlRest)

	content, err := template.New("content").Parse(string(cont))
	if err != nil {
		panic(err)
	}
	content.Execute(w, HtmlSubs)
}

func webSubscribe(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		r.ParseForm()
		// logic part of log in

		group := Subscribe{
			Time:  r.Form["time"][0],
			Users: []string{r.Form["username"][0]},
		}
		json.Marshal(group)
		b, err := json.Marshal(group)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(b)
		fmt.Println("username:", r.Form["username"])
		fmt.Println("time:", r.Form["time"])
		fmt.Println("restaurant:", r.Form["restaurant"][0]+"-"+strings.Replace(r.Form["time"][0], ":", "", -1))
		subscribe(r.Form["restaurant"][0], []byte(b))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		indexRestaurants()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func webAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		r.ParseForm()
		data := Restaurant{}
		user := r.Form["username"][0]
		json.Unmarshal([]byte(r.Form["restaurant"][0]), &data)
		fmt.Println(data.Name)
		fmt.Println(user)
		redisAppend("s/"+data.Name, ","+user)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

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

func redisGet(key string) *redis.StringCmd {
	values := redisClient.Get(key)
	return values
}

func redisAppend(key string, user string) {
	redisClient.Append(key, user)
}

func redisGetKeys(prefix string) *redis.StringSliceCmd {
	keys := redisClient.Keys(prefix)
	return keys
}
