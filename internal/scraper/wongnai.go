package scraper

import (
	"checkopen/internal/model"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/gocolly/colly/v2"
)

// WongnaiScraper handles scraping from Wongnai website
type WongnaiScraper struct {
	collector *colly.Collector
}

// NewWongnaiScraper creates a new Wongnai scraper
func NewWongnaiScraper() *WongnaiScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.wongnai.com"),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.AllowURLRevisit(), // Allow revisiting URLs for retrying requests
	)

	return &WongnaiScraper{
		collector: c,
	}
}

// GetRestaurantStatus scrapes restaurant status from Wongnai
func (ws *WongnaiScraper) GetRestaurantStatus(restaurantID string) (*model.RestaurantStatus, error) {
	url := fmt.Sprintf("https://www.wongnai.com/restaurants/%s", restaurantID)

	status := &model.RestaurantStatus{
		RestaurantID: restaurantID,
		IsOpen:       false,
		Status:       "ไม่พบข้อมูล",
	}

	var foundStatus bool
	var responseBody string

	// Capture response body
	ws.collector.OnResponse(func(r *colly.Response) {
		responseBody = string(r.Body)
	})

	ws.collector.OnHTML("body", func(e *colly.HTMLElement) {
		// Extract window._wn JSON from the response
		// Look for the pattern: window._wn = {...}
		jsonPattern := regexp.MustCompile(`window\._wn\s*=\s*({.+?});`)
		matches := jsonPattern.FindStringSubmatch(responseBody)

		if len(matches) > 1 {
			jsonStr := matches[0]

			// Parse the JSON
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr[13:len(jsonStr)-1]), &data); err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
				return
			}

			// Extract restaurant info from the store.business object
			if store, ok := data["store"].(map[string]interface{}); ok {
				if business, ok := store["business"].(map[string]interface{}); ok {

					if vv, ok := business["value"].(map[string]interface{}); ok {
						// Get restaurant name
						if name, ok := vv["name"].(string); ok {
							status.Name = name
						}
						// Get working hours status
						if whStatus, ok := vv["workingHoursStatus"].(map[string]interface{}); ok {
							if isOpen, ok := whStatus["open"].(bool); ok {
								status.IsOpen = isOpen
							}

							if status.IsOpen {
								status.Status = "เปิดอยู่"
							} else {
								status.Status = "ปิดแล้ว"
							}

							if msg, ok := whStatus["message"].(string); ok {
								status.Message = msg
								foundStatus = true
								status.OpenUntil = msg
							}

							if closing, ok := whStatus["closingSoon"].(bool); ok && closing {
								status.Status = "กำลังจะปิด"
								foundStatus = true
							}
						}
					}

				}
			}
		}
	})

	ws.collector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error scraping: %v\n", err)
	})

	err := ws.collector.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	if !foundStatus {
		status.Message = "ไม่พบข้อมูลสถานะการเปิด/ปิด"
	}

	return status, nil
}
