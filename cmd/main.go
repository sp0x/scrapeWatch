package main

import (
	"context"
	"github.com/sp0x/scrapeWatch"
)

func main() {
	_ = scrapeWatch.NonErrorStatusReceived(context.Background(), scrapeWatch.PubSubMessage{Data: []byte("testing")})
}
