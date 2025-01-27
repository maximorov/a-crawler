package crawler

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Master struct {
	domain        *url.URL
	results       map[string]int
	visited       map[string]bool
	wg            sync.WaitGroup
	linksToCrawl  chan string
	foundLinks    chan string
	errors        chan error
	workersNumber int
	stopped       atomic.Bool
}

func NewMaster(domain string, workersNumber int) (*Master, error) {
	parsedDomain, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	if parsedDomain.Scheme == "" {
		parsedDomain.Scheme = "https"
		return NewMaster(parsedDomain.String(), workersNumber)
	}
	if strings.Contains(parsedDomain.Host, `https://www`) {
		return NewMaster(strings.Replace(parsedDomain.String(), `https://www.`, `https://`, -1), workersNumber)
	}

	return &Master{
		domain:        parsedDomain,
		results:       make(map[string]int),
		visited:       make(map[string]bool),
		linksToCrawl:  make(chan string, 1000),
		foundLinks:    make(chan string, 1000),
		errors:        make(chan error, 10),
		workersNumber: workersNumber,
	}, nil
}

func (m *Master) Run(ctx context.Context) {
	if ok, err := m.isWebsiteSPA(); err != nil {
		log.Fatalln("SPA check error:", err)
	} else if ok {
		log.Println("SPA website detected")
		os.Exit(0)
	}

	for i := 0; i < m.workersNumber; i++ {
		go newCrawler(i+1, m).run(ctx)
		log.Printf("Crawler %d is started\n", i+1)
	}

	go m.handleErrors(ctx)
	go m.handleFoundLinks(ctx)

	log.Printf("Crawler is starting crawling %s\n", m.domain.String())
	m.foundLinks <- m.domain.String()

	m.wg.Add(1)
	go func() {
		<-time.After(2 * time.Second)
		m.wg.Done()
	}()

	m.wg.Wait()
}

func (m *Master) handleFoundLinks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(m.linksToCrawl)
			return
		case link := <-m.foundLinks:
			cleaned, err := m.cleanURL(link)
			if err != nil {
				m.errors <- err
				continue
			}

			if !m.visited[cleaned] {
				m.visited[cleaned] = true
				m.results[cleaned] = 1

				m.linksToCrawl <- link
			} else {
				m.results[cleaned]++
			}
		}
	}
}

func (m *Master) handleErrors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-m.errors:
			if err != nil {
				log.Println("Crawler error:", err)
			}
		}
	}
}

func (m *Master) isWebsiteSPA() (bool, error) {
	resp, err := http.Get(m.domain.String())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	spaSigns := []string{
		"react", "angular", "vue",
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	for _, sign := range spaSigns {
		if strings.Contains(strings.ToLower(string(body)), sign) {
			return true, nil
		}
	}

	return false, nil
}

func (m *Master) Results() map[string]int {
	return m.results
}

func (m *Master) Stop() {
	if m.stopped.Load() {
		return
	}
	defer m.stopped.Store(true)

	close(m.foundLinks)
	close(m.errors)
}

func (w *Master) cleanURL(rawLink string) (string, error) {
	linkURL, err := url.Parse(strings.TrimSpace(rawLink))
	if err != nil {
		return "", err
	}

	linkURL = w.domain.ResolveReference(linkURL)
	linkURL.Fragment = ""
	linkURL.RawQuery = ""
	cleaned := strings.TrimSuffix(linkURL.String(), "/")
	if cleaned == "" {
		return "", errors.New("empty url after cleaning")
	}

	return cleaned, nil
}
