// Command migrate-item-types backfills the itemTypes collection from the
// jenis values already stored on existing items.
//
// The jenis dropdown reads from the itemTypes master list (see
// GET /api/v1/items/filter-options), not from the items themselves. On a
// database that predates that collection, the master list starts empty while
// items already carry jenis values — so every existing type would silently
// disappear from the dropdown until someone re-typed it. This script closes
// that gap once.
//
// Usage:
//
//	cd backend
//	go run ./cmd/migrate-item-types            # dry run: print what would change, write nothing
//	go run ./cmd/migrate-item-types --apply    # actually insert the missing entries
//
// It only ever inserts. Existing itemTypes entries are left alone, items are
// never modified, and re-running it after an --apply is a no-op — matching
// on jenis is case-insensitive, so "Laptop" already present means "laptop"
// on an item is not inserted a second time.
package main

import (
	"context"
	"flag"
	"log"
	"sort"
	"strings"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/models"
	"my-backend/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	apply := flag.Bool("apply", false, "actually write the missing item types (default: dry run)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on already-set environment variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	client, err := db.Connect(connectCtx, db.LoadConfig())
	connectCancel()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Disconnect(disconnectCtx, client)
	}()

	itemRepo := repository.NewItemRepository(client.Database)
	itemTypeRepo := repository.NewItemTypeRepository(client.Database)

	inUse, err := itemRepo.Distinct(ctx, "jenis")
	if err != nil {
		log.Fatalf("failed to read distinct jenis from items: %v", err)
	}

	existing, err := itemTypeRepo.ListNames(ctx)
	if err != nil {
		log.Fatalf("failed to read the existing item-type master list: %v", err)
	}
	known := make(map[string]bool, len(existing))
	for _, name := range existing {
		known[strings.ToLower(strings.TrimSpace(name))] = true
	}

	missing := make([]string, 0, len(inUse))
	for _, jenis := range inUse {
		jenis = strings.TrimSpace(jenis)
		key := strings.ToLower(jenis)
		if jenis == "" || known[key] {
			continue
		}
		known[key] = true
		missing = append(missing, jenis)
	}
	sort.Strings(missing)

	log.Printf("items in use: %d distinct jenis | master list: %d entr(ies) | missing: %d\n",
		len(inUse), len(existing), len(missing))
	for _, jenis := range missing {
		log.Printf("  + %q\n", jenis)
	}

	if len(missing) == 0 {
		log.Println("nothing to do — the master list already covers every jenis in use")
		return
	}

	if !*apply {
		log.Println("dry run: nothing written. Re-run with --apply to insert the entries above.")
		return
	}

	for _, jenis := range missing {
		created, err := itemTypeRepo.Create(ctx, models.ItemType{Jenis: jenis})
		if err != nil {
			log.Fatalf("failed to insert item type %q: %v", jenis, err)
		}
		log.Printf("inserted %q (_id=%s)\n", created.Jenis, created.ID)
	}
	log.Printf("done: inserted %d item type(s)\n", len(missing))
}
