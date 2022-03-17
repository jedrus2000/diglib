package generic_library

import (
	"github.com/gocolly/colly/v2"
	"time"
)

func SetDefaultDownloadSettings(collector *colly.Collector) {
	collector.MaxBodySize = 10 * 1024 * 1024 * 1024
	collector.SetRequestTimeout(600 * time.Second)
	collector.AllowURLRevisit = true
}
