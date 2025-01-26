package crawler

import (
	"fmt"
	"sort"
)

type resultItem struct {
	URL   string
	Count int
}

func View(results map[string]int) {
	var items []resultItem
	for url, count := range results {
		items = append(items, resultItem{URL: url, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if len(items[i].URL) == len(items[j].URL) {
			return items[i].URL < items[j].URL
		}
		return len(items[i].URL) < len(items[j].URL)
	})

	for _, it := range items {
		fmt.Printf("%s (%d)\n", it.URL, it.Count)
	}
}
