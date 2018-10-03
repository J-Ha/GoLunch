package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

var query struct {
	Search struct {
		Total    graphql.Int
		Business struct {
			Name   graphql.String
			Rating graphql.Float
			Price  graphql.String
			Hours  struct {
				Is_open_now graphql.Boolean
				Open        struct {
					Start graphql.String
					End   graphql.String
				}
			}
		}
	} `graphql:"search(location: \"22765\")"`
}

func main() {
	yelpSearch()
}

func yelpSearch() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
	)

	httpClient := oauth2.NewClient(context.Background(), src)

	client := graphql.NewClient("https://api.yelp.com/v3/graphql", httpClient)

	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		// Handle error.

	}

	fmt.Println(query.Search.Business.Name)
	fmt.Println(query.Search.Total)
	fmt.Println(query.Search.Business.Rating)
	fmt.Println(query.Search.Business.Price)
	fmt.Println(query.Search.Business.Hours.Is_open_now)
}
