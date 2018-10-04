package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
	} `graphql:"search(location: $zip)"`
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

func main() {
	yelpSearch()
}

func yelpSearch() {
	variables := map[string]interface{}{
		"zip": graphql.String(os.Getenv("GOLUNCH_ZIP"))}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
	)

	httpClient := oauth2.NewClient(context.Background(), src)

	client := graphql.NewClient("https://api.yelp.com/v3/graphql", httpClient)

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		// Handle error.

	}

	fmt.Println(query.Search.Total)
	for _, business := range query.Search.Business {
		day := int(time.Now().Weekday())

		open := Open{}
		opennow := true
		if len(business.Hours) > 0 {
			openDays := business.Hours[0].Open
			open = Open{Day: int(openDays[day].Day), Start: string(openDays[day].Start), End: string(openDays[day].End)}
			opennow = bool(business.Hours[0].Is_open_now)
		}
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
		fmt.Println(string(b))
		RedisClient("r/"+string(business.Name), b)
	}

}

func RedisClient(key string, value []byte) {
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
