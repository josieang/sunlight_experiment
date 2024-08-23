package main

import (
	"context"
	"encoding/base64"
	"flag"
	"myproj/client"
	"net/http"

	"github.com/transparency-dev/merkle/proof"
	"github.com/transparency-dev/merkle/rfc6962"
)

var (
	logURL  = flag.String("log_url", "https://rome2025h2.fly.storage.tigris.dev", "log operator url without trailing slash")
	oldSize = flag.Int("old_size", 117843, "old size")
	oldHash = flag.String("old_hash", "kS5okTrSlw508uZJciXj6v2UaI2zq1bAJGiQmQVIn1E=", "old size")
)

func main() {
	ctx := context.Background()
	httpClient := &http.Client{}
	cpt, err := client.FetchCheckpoint(httpClient, *logURL)
	if err != nil {
		panic(err)
	}
	oldRoot, err := base64.StdEncoding.DecodeString(*oldHash)
	if err != nil {
		panic(err)
	}
	newRoot, err := base64.StdEncoding.DecodeString(cpt.Root)
	if err != nil {
		panic(err)
	}
	nodes, err := proof.Consistency(uint64(*oldSize), uint64(cpt.Size))
	if err != nil {
		panic(err)
	}
	cache, err := client.NewClient(httpClient, *logURL)
	if err != nil {
		panic(err)
	}
	consistencyProof := make([][]byte, len(nodes.IDs))
	for i, id := range nodes.IDs {
		h, err := cache.GetHash(ctx, uint64(id.Level), uint64(id.Index), uint64(cpt.Size))
		if err != nil {
			panic(err)
		}
		consistencyProof[i] = h[:]
	}
	rehashedProof, err := nodes.Rehash(consistencyProof, rfc6962.DefaultHasher.HashChildren)
	if err != nil {
		panic(err)
	}
	if err := proof.VerifyConsistency(rfc6962.DefaultHasher, uint64(*oldSize), uint64(cpt.Size), rehashedProof, oldRoot, newRoot); err != nil {
		panic(err)
	}
}
