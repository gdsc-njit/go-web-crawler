package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	score       int
	visitedURLs = make(map[string]bool)
	mu          sync.Mutex
	startTime   time.Time
)

func crawl(url string, depth int, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	// Avoid visiting the same URL twice
	mu.Lock()
	if visitedURLs[url] {
		mu.Unlock()
		return
	}
	visitedURLs[url] = true
	mu.Unlock()

	// Fetch the webpage
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching URL %s: %v", url, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Non-OK status code %d for URL %s", res.StatusCode, url)
		return
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Error parsing HTML for URL %s: %v", url, err)
		return
	}

	// Print the page title
	title := doc.Find("title").Text()
	fmt.Printf("Visited: %s (Title: %s)\n", url, title)

	// Update score
	mu.Lock()
	score += 10
	fmt.Printf("Score: %d\n", score)
	mu.Unlock()

	// Extract all links and crawl them
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			absoluteURL := resolveRelativeURL(url, href)
			if absoluteURL != "" {
				wg.Add(1)
				go crawl(absoluteURL, depth-1, wg)
			}
		}
	})
}

func resolveRelativeURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}
	if strings.HasPrefix(relativeURL, "/") {
		return baseURL + relativeURL
	}
	return ""
}

func main() {
	startURL := "https://go.dev" // Starting URL
	maxDepth := 2                     // Maximum depth to crawl
	startTime = time.Now()            // Start the timer

	var wg sync.WaitGroup
	wg.Add(1)
	go crawl(startURL, maxDepth, &wg)
	wg.Wait()

	// Calculate time taken
	elapsed := time.Since(startTime)
	fmt.Printf("Crawl completed in %s. Final Score: %d\n", elapsed, score)
}