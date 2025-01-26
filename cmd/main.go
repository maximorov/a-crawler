package main

import (
	"flag"
	"fmt"
	"github.com/maximorov/a-crawler/internal/pkg"
	"log"
	"os"
)

func main() {
	threads := flag.Int("threads", 4, "Threads number (default if 4)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s [--threads N] <domain>\n", os.Args[0])
		os.Exit(1)
	}

	domain := args[0]
	maxWorkers := *threads

	crawler, err := pkg.NewCrawler(domain, maxWorkers)
	if err != nil {
		log.Fatalf("Помилка створення краулера: %v", err)
	}
}
