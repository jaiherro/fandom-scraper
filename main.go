package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	downloadQueue := make(chan []string, 500000)

	var links [][]string

	start := time.Now()
	fmt.Println("[INFO] Starting link aggregation")
	links = append(links, looneyScraper()...)
	links = append(links, simpsonsScraper()...)
	links = append(links, spongebobScraper()...)
	fmt.Printf("[INFO] Link aggregation completed in %s\n", time.Since(start))

	fmt.Printf("[INFO] Total links collected %d\n", len(links))

	go func() {
		for _, link := range links {
			downloadQueue <- link
		}
	}()

	cwd, _ := os.Getwd()
	os.Mkdir(filepath.Join(cwd+"/"+"Simpsons"), os.ModePerm)
	os.Mkdir(filepath.Join(cwd+"/"+"Spongebob"), os.ModePerm)
	os.Mkdir(filepath.Join(cwd+"/"+"Looney Tunes"), os.ModePerm)

	var wg sync.WaitGroup

	wg.Add(len(links))

	for i := 1; i < 25; i++ {
		go downloader(downloadQueue, &wg)
	}

	wg.Wait()
	fmt.Println("[INFO] Queue complete, destroying agents and exiting")
	close(downloadQueue)
	fmt.Println("[INFO] Goodbye")
}

func simpsonsScraper() [][]string {
	fmt.Println("[INFO] Starting Simpsons scrape")
	start := time.Now()
	var links [][]string

	threadCollector := colly.NewCollector(
		colly.Async(true),
	)
	imageCollector := colly.NewCollector(
		colly.Async(true),
	)

	threadCollector.OnHTML("tbody tr td.oLeft b a", func(e *colly.HTMLElement) {
		imageCollector.Visit("https://simpsons.fandom.com" + e.Attr("href") + "/Gallery")
	})

	imageCollector.OnHTML("a.image.lightbox img", func(e *colly.HTMLElement) {
		links = append(links, []string{(strings.Split(e.Attr("data-src"), "/revision/")[0]), e.Attr("data-image-name"), "Simpsons"})
	})

	threadCollector.Visit("https://simpsons.fandom.com/wiki/List_of_Episodes")
	threadCollector.Wait()
	imageCollector.Wait()
	fmt.Printf("[INFO] Simpsons scrape completed in %s\n", time.Since(start))
	return links
}

func spongebobScraper() [][]string {
	fmt.Println("[INFO] Starting Spongebob scrape")
	start := time.Now()
	var links [][]string

	threadCollector := colly.NewCollector(
		colly.Async(true),
	)
	imageCollector := colly.NewCollector(
		colly.Async(true),
	)

	threadCollector.OnHTML("tbody tr td[style='text-align:left'] a", func(e *colly.HTMLElement) {
		imageCollector.Visit("https://spongebob.fandom.com" + e.Attr("href") + "/gallery")
	})

	imageCollector.OnHTML("a.image.lightbox img", func(e *colly.HTMLElement) {
		links = append(links, []string{(strings.Split(e.Attr("data-src"), "/revision/")[0]), e.Attr("data-image-name"), "Spongebob"})
	})

	threadCollector.Visit("https://spongebob.fandom.com/wiki/List_of_Episodes")
	threadCollector.Wait()
	imageCollector.Wait()
	fmt.Printf("[INFO] Spongebob scrape completed in %s\n", time.Since(start))
	return links
}

func looneyScraper() [][]string {
	fmt.Println("[INFO] Starting Looney Tunes scrape")
	start := time.Now()
	var links [][]string

	threadCollector := colly.NewCollector(
		colly.Async(true),
	)
	imageCollector := colly.NewCollector(
		colly.Async(true),
	)

	threadCollector.OnHTML("table > tbody > tr > td > table > tbody > tr > td > div > div > div > a", func(e *colly.HTMLElement) {
		imageCollector.Visit("https://looneytunes.fandom.com" + e.Attr("href") + "/Gallery")
		imageCollector.Visit("https://looneytunes.fandom.com" + e.Attr("href"))
	})

	imageCollector.OnHTML("a.image.lightbox img", func(e *colly.HTMLElement) {
		links = append(links, []string{(strings.Split(e.Attr("data-src"), "/revision/")[0]), e.Attr("data-image-name"), "Looney Tunes"})
	})

	threadCollector.Visit("https://looneytunes.fandom.com/wiki/Main_Page")
	threadCollector.Wait()
	imageCollector.Wait()
	fmt.Printf("[INFO] Looney Tunes scrape completed in %s\n", time.Since(start))
	return links
}

func downloader(downloadQueue <-chan []string, wg *sync.WaitGroup) {
	reg, _ := regexp.Compile("[^a-zA-Z0-9\\.]+")
	cwd, _ := os.Getwd()
	for entry := range downloadQueue {
		fmt.Printf("[DOWN] %s\n", entry[1])
		response, e := http.Get(entry[0])
		if e != nil {
			log.Fatal(e)
		}

		if response.StatusCode == http.StatusOK {
			var file *os.File

			for {
				entry[1] = reg.ReplaceAllString(entry[1], "")
				if entry[1] != "" {
					break
				} else {
					entry[1] = strconv.Itoa(rand.Intn(999))
				}
			}

			if _, err := os.Stat(entry[1]); err == nil {
				splitname := strings.Split(entry[1], ".")
				var name string
				for {
					name = splitname[0] + strconv.Itoa(rand.Intn(999)) + "." + splitname[len(splitname)-1]
					if _, err := os.Stat(name); os.IsNotExist(err) {
						break
					}
				}
				file, err = os.Create(filepath.Join(cwd+"/"+entry[2], filepath.Base(name)))
				if err != nil {
					log.Fatal(err)
				}
			} else if os.IsNotExist(err) {
				file, err = os.Create(filepath.Join(cwd+"/"+entry[2], filepath.Base(entry[1])))
				if err != nil {
					log.Fatal(err)
				}
			}
			_, err := io.Copy(file, response.Body)
			if err != nil {
				log.Fatal(err)
			}

			file.Close()
		} else {
			fmt.Println("[WARN] Download declined by server, dropping link")
		}
		response.Body.Close()
		wg.Done()
	}
}
