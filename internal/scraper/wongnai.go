package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"checkopen/internal/model"

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
		closingSoonPattern := regexp.MustCompile(`฿฿กำลังจะปิด`)
		closedPattern := regexp.MustCompile(`฿฿ปิดอยู่`)

		if matches := openPattern.FindStringSubmatch(bodyText); len(matches) > 1 && matches[1] != "" {
			status.IsOpen = true
			status.Status = "เปิดอยู่"
			foundStatus = true
			status.OpenUntil = matches[1]
			status.Message = fmt.Sprintf("ร้านเปิดอยู่จนถึง %s น.", matches[1])
		} else if matches := closingSoonPattern.FindStringSubmatch(bodyText); len(matches) > 0 {
			status.IsOpen = true
			status.Status = "กำลังจะปิด"
			foundStatus = true
			if len(matches) > 1 && matches[1] != "" {
				status.OpenUntil = matches[1]
				status.Message = fmt.Sprintf("ร้านกำลังจะปิด (จนถึง %s น.)", matches[1])
			} else {
				status.Message = "ร้านกำลังจะปิด"
			}
		} else if closedPattern.MatchString(bodyText) {
			status.IsOpen = false
			status.Status = "ปิด"
			status.Message = "ร้านปิดอยู่"
			foundStatus = true
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
