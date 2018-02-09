//Go has built-in concurrency which we are using to speedup the fetching time
//adding concurrency is to speed up the loading
//so what is basically happening is that we send a request and then wait for it to respond 
//go routines provide concurrency
//go routine is like a light weight thread
//we need to synchronize
//we can use channels, defer other stuffs

package main
// := for assigning necessary for initialization, rest all places can work with =
// var grades ... can be written as grades:= make(.....) if inside a function

import (
	"fmt"
	"net/http"
	"html/template"
	"encoding/xml"
	"io/ioutil"
	"sync"
)

var wg sync.WaitGroup

type NewsMap struct {
	Keyword string
	Location string
}

type NewsAggPage struct {
    Title string
    News map[string]NewsMap
}

type Sitemapindex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles []string `xml:"url>news>title"`
	Keywords []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Whoa, Go is neat!</h1>")
}

//this function below is to add routines
func newsRoutine (c chan News,Location string){

	defer wg.Done()
	var n News
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)					//this loop is really slow because its going to the site getting sitemap going to next and next
	resp.Body.Close()

	c<- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s Sitemapindex
	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	news_map := make(map[string]NewsMap)
	resp.Body.Close()				//check this out
	//string_body := string(bytes)
	//fmt.Println(string_body)
	//resp.Body.Close()

	queue := make(chan News, 30)
	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue,Location)	

	}

	wg.Wait()
	close(queue)

	//the reason we removed it from the before loop is that we need to access bunch in form of channels
	for elem := range queue {
		for idx, _ := range elem.Keywords {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "Amazing News Aggregator", News: news_map}	//passing the news_map to News variable
    t, _ := template.ParseFiles("newsaggtemplate.html")					//giving the template
    t.Execute(w, p)														//executing the function to display
    
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/agg/", newsAggHandler)
	http.ListenAndServe(":8000", nil) 
}
