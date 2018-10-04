package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
	} `graphql:"search(location: $zip, limit: 1, radius: 1500)"`
}

type Restaurant struct {
	Name      string
	URL       string
	Price     string
	IsOpenNow bool
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
	for num := range query.Search.Business {
		fmt.Println(query.Search.Business[num].Name)
		fmt.Println(query.Search.Business[num].Url)
		fmt.Println(query.Search.Business[num].Rating)
		fmt.Println(query.Search.Business[num].Price)
		fmt.Println(query.Search.Business[num].Hours[0].Is_open_now)
		m := Restaurant{string(query.Search.Business[num].Name), string(query.Search.Business[num].Url), string(query.Search.Business[num].Price), bool(query.Search.Business[num].Hours[0].Is_open_now)}
		b, err := json.Marshal(m)
		if err != nil {
			// Handle error.
		}
		fmt.Println(b)
		for day := range query.Search.Business[num].Hours[0].Open {
			fmt.Println(query.Search.Business[num].Hours[0].Open[day].Day)
			fmt.Println(query.Search.Business[num].Hours[0].Open[day].Start)
			fmt.Println(query.Search.Business[num].Hours[0].Open[day].End)
		}

	}

}
