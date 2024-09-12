package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"

	"myproj/client"

	"github.com/transparency-dev/merkle/rfc6962"
)

var (
	logURL = flag.String("log_url", "https://rome2025h2.fly.storage.tigris.dev", "log url without trailing slash")
)

func main() {
	ctx := context.Background()
	httpClient := &http.Client{}
	cpt, err := client.FetchCheckpoint(httpClient, *logURL)
	if err != nil {
		panic(err)
	}
	fmt.Printf("checkpoint size: %d\n", cpt.Size)
	c, err := client.NewClient(httpClient, *logURL)
	if err != nil {
		panic(err)
	}

	// Calculate hashes from entries.
	entryTile, err := c.GetEntryTile(ctx, 0, uint64(cpt.Size))
	if err != nil {
		panic(err)
	}
	entryTileHashes := make([][sha256.Size]byte, 256)
	for i, e := range entryTile.Entries {
		copy(entryTileHashes[i][:], rfc6962.DefaultHasher.HashLeaf(e.LeafInput))
	}

	// Fetch the same hashes from the tile endpoint.
	tile, err := c.GetTile(ctx, 0, 0, uint64(cpt.Size))
	if err != nil {
		panic(err)
	}
	tileHashes := tile.Nodes[0]

	// Chck that the hashes are the same.
	for i := range entryTileHashes {
		if eh, h := entryTileHashes[i], tileHashes[i]; eh != h {
			fmt.Printf("%d is different entryTile/tile: %s/%s\n", i, hex.EncodeToString(eh[:])[:10], hex.EncodeToString(h[:])[:10])
		}
	}

}
