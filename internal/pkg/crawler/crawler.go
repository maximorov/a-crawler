package crawler

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type crawler struct {
	id     int
	master *Master
	client *http.Client
}

func newCrawler(id int, master *Master) *crawler {
	return &crawler{
		id:     id,
		master: master,
		client: &http.Client{},
	}
}

func (w *crawler) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case link := <-w.master.linksToCrawl:
			w.master.wg.Add(1)
			defer w.master.wg.Done()

			log.Printf("Crawler %d is crawling %s\n", w.id, link)
			if err := w.crawl(link); err != nil {
				w.master.errors <- err
			}
		}
	}
}

func (w *crawler) crawl(link string) error {
	resp, err := w.client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("status code is %s", resp.Status))
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if w.isSameDomain(attr.Val) {
						w.master.foundLinks <- attr.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return nil
}

func (w *crawler) isSameDomain(linkStr string) bool {
	u, err := url.Parse(linkStr)
	if err != nil {
		return false
	}
	return strings.HasSuffix(u.Host, w.master.domain.Hostname())
}
