// Package for working with RSS-feeds
package rss

import (
	"GoNews/news/pkg/storage"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
)

type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Chanel  Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

// Parse reads the rss stream and returns an array of decoded news.
func Parse(url string) ([]newsStorage.Post, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var f Feed
	err = xml.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}
	var data []newsStorage.Post
	for _, item := range f.Chanel.Items {
		var p newsStorage.Post
		p.Title = item.Title
		p.Content = item.Description
		p.Content = strip.StripTags(p.Content)
		p.Link = item.Link
		item.PubDate = strings.ReplaceAll(item.PubDate, ",", "")
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", item.PubDate)
		}
		if err == nil {
			p.PubTime = t.Unix()
		}
		data = append(data, p)
	}
	return data, nil
}
