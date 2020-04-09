package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("godoc.org"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("table.wikitable tbody tr td b a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasSuffix(link, "jpg") || strings.HasSuffix(link, "png") {
			download(link)
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	c.Visit("https://simpsons.fandom.com/wiki/List_of_Episodes/")
}

func download(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(ioutil.ReadAll(resp.Body))
}
