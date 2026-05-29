package rss

import (
	"context"
	"encoding/xml"
	"html"
	"net/http"
)

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Item        []RSSItem `xml:"item"`
}

type RSSFeed struct {
	Channel RSSChannel `xml:"channel"`
}

func (c *RSSChannel) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias RSSChannel
	var a Alias

	if err := d.DecodeElement(&a, &start); err != nil {
		return err
	}

	*c = RSSChannel(a)
	c.Title = html.UnescapeString(c.Title)
	c.Description = html.UnescapeString(c.Description)
	return nil
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (i *RSSItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias RSSItem
	var a Alias

	if err := d.DecodeElement(&a, &start); err != nil {
		return err
	}

	*i = RSSItem(a)

	i.Title = html.UnescapeString(i.Title)
	i.Description = html.UnescapeString(i.Description)

	return nil
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	c := new(http.Client)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	feed := new(RSSFeed)
	xml.NewDecoder(res.Body).Decode(feed)

	return feed, nil
}
