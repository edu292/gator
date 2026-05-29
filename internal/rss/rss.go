package rss

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"net/http"
	"time"

	"gator/internal/database"

	"github.com/lib/pq"
)

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
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
	PubDate     *time.Time
}

func (i *RSSItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		PubDate     string `xml:"pubDate"`
	}
	var a Alias

	if err := d.DecodeElement(&a, &start); err != nil {
		return err
	}

	i.Title = html.UnescapeString(a.Title)
	i.Link = a.Link
	i.Description = html.UnescapeString(a.Description)

	if a.PubDate != "" {
		var parsedTime time.Time
		var err error

		parsedTime, err = time.Parse(time.RFC1123Z, a.PubDate)
		if err != nil {
			parsedTime, err = time.Parse(time.RFC1123, a.PubDate)
		}

		if err == nil && !parsedTime.IsZero() {
			i.PubDate = &parsedTime
		}
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
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
	if err := xml.NewDecoder(res.Body).Decode(feed); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	return feed, nil
}

func ScrapeFeeds(ctx context.Context, db *database.Queries) error {
	feedDB, err := db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("could not get next feed to fetch: %w", err)
	}

	err = db.MarkFeedFetched(ctx, feedDB.ID)
	if err != nil {
		return fmt.Errorf("could not mark the feed as fetched: %s", err)
	}

	feed, err := fetchFeed(ctx, feedDB.Url)
	if err != nil {
		return fmt.Errorf("could not fetch field: %w", err)
	}

	for _, item := range feed.Channel.Items {
		var publishedAtSQL sql.NullTime
		if item.PubDate != nil {
			publishedAtSQL = sql.NullTime{
				Time:  *item.PubDate,
				Valid: true,
			}
		}
		_, err := db.CreatePost(ctx, database.CreatePostParams{
			PublishedAt: publishedAtSQL,
			Title:       item.Title,
			Description: item.Description,
			Url:         item.Link,
			FeedID:      feedDB.ID,
		})
		if err != nil {
			if pgErr, ok := errors.AsType[*pq.Error](err); ok {
				if pgErr.Code != "23505" {
					return err
				}
			}
		}
	}
	return nil
}
