package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"myproj/client"

	"github.com/transparency-dev/merkle/rfc6962"
)

var (
	logURL         = flag.String("log_url", "https://rome2025h2.fly.storage.tigris.dev", "log url without trailing slash")
	tileIndex      = flag.Int("tile_index", 460, "tile to check")
	hashIndex      = flag.Int("hash_index", 117844, "specific hash to check")
	hashRangeIndex = flag.String("hash_range", "117760,117844", "hash range to check")
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

	if *tileIndex != -1 {
		// Calculate hashes from entries.
		entryTile, err := c.GetEntryTile(ctx, uint64(*tileIndex), uint64(cpt.Size))
		if err != nil {
			panic(err)
		}
		entryTileHashes := make([][sha256.Size]byte, len(entryTile.Entries))
		for i, e := range entryTile.Entries {
			copy(entryTileHashes[i][:], rfc6962.DefaultHasher.HashLeaf(e.LeafInput))
		}

		// Fetch the same hashes from the tile endpoint.
		tile, err := c.GetTile(ctx, 0, uint64(*tileIndex), uint64(cpt.Size))
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

	if *hashIndex != -1 {
		// Calculate hashes from entries.
		entries, err := c.GetEntries(ctx, uint64(*hashIndex), uint64(*hashIndex), uint64(cpt.Size))
		if err != nil {
			panic(err)
		}
		var entryHash [sha256.Size]byte
		copy(entryHash[:], rfc6962.DefaultHasher.HashLeaf(entries.Entries[0].LeafInput))

		// Fetch the same hashes from the tile endpoint.
		hash, err := c.GetHash(ctx, 0, uint64(*hashIndex), uint64(cpt.Size))
		if err != nil {
			panic(err)
		}
		if entryHash != hash {
			fmt.Printf("hash %d is different entryTile/tile: %s/%s\n", *hashIndex, hex.EncodeToString(entryHash[:])[:10], hex.EncodeToString(hash[:])[:10])
		}
	}

	if *hashRangeIndex != "" {
		parts := strings.Split(*hashRangeIndex, ",")
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}

		// Calculate hashes from entries.
		entries, err := c.GetEntries(ctx, uint64(start), uint64(end), uint64(cpt.Size))
		if err != nil {
			panic(err)
		}
		entryTileHashes := make([][sha256.Size]byte, end-start+1)
		for i := 0; i < len(entryTileHashes); i++ {
			copy(entryTileHashes[i][:], rfc6962.DefaultHasher.HashLeaf(entries.Entries[i].LeafInput))
		}

		// Fetch the same tileHashes from the tile endpoint.
		tileHashes := make([][sha256.Size]byte, end-start+1)
		i := 0
		for index := start; index <= end; index++ {
			tileHashes[i], err = c.GetHash(ctx, 0, uint64(index), uint64(cpt.Size))
			if err != nil {
				panic(err)
			}
			i += 1
		}
		// Chck that the hashes are the same.
		for i := range entryTileHashes {
			if eh, h := entryTileHashes[i], tileHashes[i]; eh == h {
				fmt.Printf("%d is different entryTile/tile: %s/%s\n", i, hex.EncodeToString(eh[:])[:10], hex.EncodeToString(h[:])[:10])
			}
		}
	}

}
