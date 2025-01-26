package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/maximorov/a-crawler/internal/pkg/crawler"
	"log"
	"os"
	"os/signal"
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
	workersNum := *threads

	processor, err := crawler.NewMaster(domain, workersNum)
	if err != nil {
		log.Fatalf("Crawler creating error: %v", err)
	}
	log.Printf("Crawling %s with %d workers\n", domain, workersNum)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	processor.Run(ctx)

	crawler.View(processor.Results())
}
