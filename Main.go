package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
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

type DropdownItem struct {
	Name  string
	Value string
}

func main() {
	user := getUser()
	restaurant := ("Asia Imbiss")
	subscribe(restaurant, user)
	http.HandleFunc("/", generateWebsite)
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
			redisClient("r/"+string(business.Name), b)
		}
		offset = offset + limit
	}
}

func redisClient(key string, value []byte) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := client.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func redisGet(key string) *redis.StringCmd {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	values := client.Get(key)
	return values
}

func redisGetKeys(prefix string) *redis.StringSliceCmd {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	keys := client.Keys(prefix)
	return keys
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
	redisClient("s/"+restaurant, user)
}

func generateWebsite(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
	<html>
	<body>
   
	<select> // for loop in html template example
	  {{range $key, $value := .}}
		<option value="{{ $value }}">{{ $key }}</option>
	  {{end}}
   
   
	</select>
   
	</body>
	</html>`

	names, _ := redisGetKeys("r/*").Result()

	for num, rest := range names {
		values, _ := redisGet(rest).Result()
		var htmlrest = map[string]interface{}{
			strings.Replace(names[num], "r/", "", -1): values,
		}
		//io.WriteString(w, strings.Replace(names[num], "r/", "", -1)+" "+values+"\n")
		dropdownTemplate, err := template.New("dropdownexample").Parse(string(html))
		if err != nil {
			panic(err)
		}
		// populate dropdown with fruits
		dropdownTemplate.Execute(w, htmlrest)

	}

}
