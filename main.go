package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
)

type CmdOpts struct {
	URL            string
	AllowedDomains string
}

func main() {
	opts := &CmdOpts{}

	f := flag.NewFlagSet("", flag.ExitOnError)
	f.StringVar(&opts.URL, "url", "", "the url to check")
	f.StringVar(&opts.AllowedDomains, "allowed-domains", "", "list of allowed domains to crawl")
	err := f.Parse(os.Args[1:])
	if err != nil {
		log.Fatal("failed to parse arguments")
	}

	var urlsChecked int
	var urlsFailed int
	knownUrls := make(map[string]string)

	q, _ := queue.New(1, &queue.InMemoryQueueStorage{
		MaxSize: 10000,
	})

	q.AddURL(opts.URL)

	c := colly.NewCollector(
		colly.AllowedDomains(strings.Split(opts.AllowedDomains, ",")...),
	)

	c.OnRequest(func(r *colly.Request) {
		urlsChecked++
		fmt.Printf("Checking: %s\n", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		urlsFailed++
		fmt.Printf("FAILED: %s (%d)\nERROR: %+v\n", r.Request.URL, r.StatusCode, err)
	})

	c.OnHTML("a[href]", func(h *colly.HTMLElement) {
		link := h.Attr("href")

		if !strings.HasPrefix(link, "http") {
			return
		}

		if _, exist := knownUrls[link]; !exist {
			knownUrls[link] = h.Text
			q.AddURL(link)
		}
	})

	q.Run(c)

	fmt.Println("All known URLs:")
	for link, text := range knownUrls {
		fmt.Printf("\t- %q -> %s\n", text, link)
	}

	fmt.Println("Checked", urlsChecked)
	fmt.Println("Failed", urlsFailed)
	fmt.Println("Collected", len(knownUrls), "URLs")
}
