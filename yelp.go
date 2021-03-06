package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"os"
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
			if redisListLength(string(business.Name)) == 0 {
				redisSetList(string(business.Name), string(b))
			}
		}
		offset = offset + limit
	}
}
