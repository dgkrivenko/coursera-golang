package main

import (
	"context"
	"log"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	err := StartMyMicroservice(ctx, "127.0.0.1:8082", "")
	if err != nil {
		log.Fatalf("cant start server initial: %v", err)
	}
	println("usage: go test -v")
}
