package pkg

import (
	"net/url"
)

type Crawler struct {
	domain *url.URL
}

func NewCrawler(domainStr string, maxWorkers int) (*Crawler, error) {
	parsedDomain, err := url.Parse(domainStr)
	if err != nil {
		return nil, err
	}

	if parsedDomain.Scheme == "" {
		parsedDomain.Scheme = "https"
	}

	c := &Crawler{
		domain: parsedDomain,
	}
	return c, nil
}
