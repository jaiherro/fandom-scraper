package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	threadCollector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"),
	)
	imageCollector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"),
	)

	threadCollector.OnHTML("tbody tr td.oLeft b a", func(e *colly.HTMLElement) {
		imageCollector.Visit("https://simpsons.fandom.com" + e.Attr("href") + "/Gallery")
	})

	imageCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	imageCollector.OnHTML("a.image.lightbox img", func(e *colly.HTMLElement) {
		go download(strings.Split(e.Attr("data-src"), "/revision/")[0], e.Attr("data-image-name"))
		time.Sleep(200 * time.Millisecond)
	})

	threadCollector.Visit("https://simpsons.fandom.com/wiki/List_of_Episodes")
}

func download(url string, name string) {
	fmt.Println("Downloading:", name)
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
}
