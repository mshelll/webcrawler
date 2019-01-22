package main

import (
	"fmt"
	"sync"
	"time"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, crawl_info CrawlInfo) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := crawl_info.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		go Crawl(u, depth-1, crawl_info)
	}
	return
}

func main() {
	go Crawl("https://golang.org/", 4, crawl_info)
	
	time.Sleep(time.Second)
}

type CrawlInfo struct {
	fetcher fakeFetcher
	cache   urlCache
	mux     sync.Mutex
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult
type urlCache map[string]struct{}

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

func (crawl_info CrawlInfo) Fetch(url string) (string, []string, error) {
	crawl_info.mux.Lock()
	
	if _, ok := crawl_info.cache[url]; ok {
		return "", nil, fmt.Errorf("url already found: %s", url)
	}
	if res, ok := crawl_info.fetcher[url]; ok {
		
		crawl_info.cache[url] = struct{}{}
		return res.body, res.urls, nil
	}
	crawl_info.mux.Unlock()
	return "", nil, fmt.Errorf("not found: %s", url)
}

var crawl_info CrawlInfo

func init() {

	// fetcher is a populated fakeFetcher.
	crawl_info.fetcher = map[string]*fakeResult{
		"https://golang.org/": &fakeResult{
			"The Go Programming Language",
			[]string{
				"https://golang.org/pkg/",
				"https://golang.org/cmd/",
			},
		},
		"https://golang.org/pkg/": &fakeResult{
			"Packages",
			[]string{
				"https://golang.org/",
				"https://golang.org/cmd/",
				"https://golang.org/pkg/fmt/",
				"https://golang.org/pkg/os/",
			},
		},
		"https://golang.org/pkg/fmt/": &fakeResult{
			"Package fmt",
			[]string{
				"https://golang.org/",
				"https://golang.org/pkg/",
			},
		},
		"https://golang.org/pkg/os/": &fakeResult{
			"Package os",
			[]string{
				"https://golang.org/",
				"https://golang.org/pkg/",
			},
		},
	}
	
	crawl_info.cache = make(map[string]struct{})

}
