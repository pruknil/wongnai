package scraper

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"checkopen/internal/model"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// WongnaiScraper handles scraping from Wongnai website
type WongnaiScraper struct {
	collector    *colly.Collector
	maxRetries   int
	retryDelay   time.Duration
	requestDelay time.Duration
}

// NewWongnaiScraper creates a new Wongnai scraper
func NewWongnaiScraper() *WongnaiScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.wongnai.com"),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.AllowURLRevisit(), // Allow revisiting URLs for retrying requests
	)

	// Set rate limiting: 2 seconds delay between requests
	c.Limit(&colly.LimitRule{
		Parallelism: 1,
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	// Set request timeout
	c.SetRequestTimeout(15 * time.Second)

	// Add better headers to avoid rate limiting
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "th-TH,th;q=0.9,en;q=0.8")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Referer", "https://www.wongnai.com/")
	})

	// Enable verbose logging for debugging
	c.SetDebugger(&debug.LogDebugger{})

	return &WongnaiScraper{
		collector:    c,
		maxRetries:   3,
		retryDelay:   2 * time.Second,
		requestDelay: 2 * time.Second,
	}
}

// GetRestaurantStatus scrapes restaurant status from Wongnai with retry logic
func (ws *WongnaiScraper) GetRestaurantStatus(restaurantID string) (*model.RestaurantStatus, error) {
	url := fmt.Sprintf("https://www.wongnai.com/delivery/businesses/%s/order", restaurantID)

	status := &model.RestaurantStatus{
		RestaurantID: restaurantID,
		IsOpen:       false,
		Status:       "ไม่พบข้อมูล",
	}

	var lastErr error
	var foundStatus bool

	// Retry logic with exponential backoff
	for attempt := 0; attempt <= ws.maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff: 2^attempt seconds with some jitter
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			fmt.Printf("Retry attempt %d/%d after %v...\n", attempt, ws.maxRetries, backoff)
			time.Sleep(backoff)
		}

		foundStatus = false
		lastErr = nil

		// Set up the OnHTML handler
		ws.collector.OnHTML("body", func(e *colly.HTMLElement) {
			// ค้นหาชื่อร้าน
			e.ForEach("h1", func(_ int, el *colly.HTMLElement) {
				if status.Name == "" {
					status.Name = strings.TrimSpace(el.Text)
				}
			})

			// ค้นหาคำว่า "เปิดอยู่" ในเนื้อหาทั้งหมด
			bodyText := e.Text

			// Pattern สำหรับหาสถานะเปิด/ปิด
			openPattern := regexp.MustCompile(`เปิดอยู่(?:จนถึง\s*(\d{1,2}:\d{2}))?`)
			closedPattern := regexp.MustCompile(`ปิดอยู่`)

			if matches := openPattern.FindStringSubmatch(bodyText); len(matches) > 1 && matches[1] != "" {
				status.IsOpen = true
				status.Status = "เปิดอยู่"
				foundStatus = true
				status.OpenUntil = matches[1]
				status.Message = fmt.Sprintf("ร้านเปิดอยู่จนถึง %s น.", matches[1])
			} else if closedPattern.MatchString(bodyText) {
				status.IsOpen = false
				status.Status = "ปิด"
				status.Message = "ร้านปิดอยู่"
				foundStatus = true
			}
		})

		// Set up the OnError handler
		ws.collector.OnError(func(r *colly.Response, err error) {
			lastErr = err
			fmt.Printf("Error scraping (attempt %d): %v (status code: %d)\n", attempt+1, err, r.StatusCode)
		})

		// Attempt to visit the URL
		err := ws.collector.Visit(url)
		if err != nil && lastErr == nil {
			lastErr = err
		}

		// If successful, return
		if foundStatus || err == nil {
			break
		}

		// Check if it's a rate limiting error
		if lastErr != nil && strings.Contains(lastErr.Error(), "429") {
			// For 429 Too Many Requests, increase backoff
			fmt.Printf("Rate limited, increasing backoff...\n")
			continue
		}

		// If we got an actual response, don't retry
		if err == nil && !foundStatus {
			break
		}
	}

	// Clear handlers to prevent memory issues
	ws.collector.OnHTML("body", nil)
	ws.collector.OnError(nil)

	if lastErr != nil && !foundStatus {
		return nil, fmt.Errorf("failed to scrape after %d attempts: %w", ws.maxRetries+1, lastErr)
	}

	if !foundStatus {
		status.Message = "ไม่พบข้อมูลสถานะการเปิด/ปิด"
	}

	return status, nil
}
