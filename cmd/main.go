package main

import (
	"context"
	"scrapeWatch"
)

func main() {
	scrapeWatch.NonErrorStatusReceived(context.Background(), scrapeWatch.PubSubMessage{Data: []byte("testing")})
}
