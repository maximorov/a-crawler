package crawler

import (
	"context"
	"github.com/gocolly/colly"
	"log"
	"net/url"
	"strings"
)

type crawler struct {
	id     int
	master *Master
	cl     *colly.Collector
}

func newCrawler(id int, master *Master) *crawler {
	return &crawler{
		id:     id,
		master: master,
		cl:     colly.NewCollector(),
	}
}

func (w *crawler) run(ctx context.Context) {
	w.init()

	for {
		select {
		case <-ctx.Done():
			return
		case link := <-w.master.linksToCrawl:
			w.master.wg.Add(1)
			defer w.master.wg.Done()

			w.master.errors <- w.cl.Visit(link)
		}
	}
}

func (w *crawler) init() {
	w.cl.OnRequest(func(r *colly.Request) {
		log.Printf("Crawler %d is crawling %s\n", w.id, r.URL.String())
	})
	w.cl.OnError(func(_ *colly.Response, err error) {
		w.master.errors <- err
	})
	w.cl.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if link, ok := w.checkDomain(e.Attr("href")); ok {
			w.master.foundLinks <- link
		}
	})
}

func (w *crawler) checkDomain(linkStr string) (string, bool) {
	u, err := url.Parse(linkStr)
	if err != nil {
		return ``, false
	}

	if u.Host == `` && strings.Contains(u.Path, `/`) {
		u.Scheme = w.master.domain.Scheme
		u.Host = w.master.domain.Hostname()
	}

	return u.String(), strings.HasSuffix(u.Host, w.master.domain.Hostname())
}
