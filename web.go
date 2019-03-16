package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

func webIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		indexRestaurants()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateWebsite(w http.ResponseWriter, r *http.Request) {
	html, _ := ioutil.ReadFile("template.html")
	cont, _ := ioutil.ReadFile("content.html")
	names := redisGetKeys("*")

	var HtmlRest = make(map[string]interface{})
	var HtmlSubs = make(map[string]interface{})
	//var HtmlTimes = make(map[string]interface{})
	for num, rest := range names {
		HtmlRest[names[num]] = rest

		if redisListLength(names[num]) >= 2 {
			HtmlSubs[redisGetList(rest, 0, 1)[0]] = redisGetList(rest, 1, redisListLength(names[num]))

		}

		//HtmlTimes[]
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
		group := Subscribe{
			Time:  r.Form["time"][0],
			Users: []string{r.Form["username"][0]},
		}

		json.Marshal(group)
		b, err := json.Marshal(group)
		if err != nil {
			fmt.Println("error:", err)
		}
		subscribe(r.Form["restaurant"][0], []byte(b))
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
