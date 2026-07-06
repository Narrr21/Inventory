// Command seed populates MongoDB with a handful of sample items for local
// development and manual testing.
//
// Usage:
//
//	cd backend
//	go run ./cmd/seed          # insert the sample items (additive)
//	go run ./cmd/seed --reset  # delete all existing items first, then insert
//
// It connects using the same MONGODB_URI / MONGODB_DB as cmd/server (loaded
// from backend/.env if present), so make sure that's set up first.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/models"
	"my-backend/internal/repository"

	"github.com/joho/godotenv"
)

// seedItems mirrors the old Phase 1 mock fixture data, for local/manual testing.
// Status values match the frontend's canonical set: Healthy | Under Maintenance | Broken.
var seedItems = []models.Item{
	{
		JenisProduct: "Laptop", SerialNumber: "SN-00123", Name: "RTI-ALPHA-001", Proyek: "ALPHA",
		PasswordPin: "1234", Account: "user01", PasswordAccount: "secret01", IPAddress: "10.0.0.12",
		Anydesk: "123 456 789", Rustdesk: "", PasswordAnydeskRustdesk: "rdpass01",
		LicenseWindows: "Pro", LicenseOffice: "365", Status: "Healthy", Lokasi: "Jakarta HQ",
		Deskripsi: "Contoh data seed",
	},
	{
		JenisProduct: "PC", SerialNumber: "SN-00124", Name: "RTI-BETA-002", Proyek: "BETA",
		PasswordPin: "5678", Account: "user02", PasswordAccount: "secret02", IPAddress: "10.0.0.13",
		Anydesk: "234 567 890", Rustdesk: "rd-002", PasswordAnydeskRustdesk: "rdpass02",
		LicenseWindows: "Home", LicenseOffice: "2021", Status: "Under Maintenance", Lokasi: "Surabaya Branch",
		Deskripsi: "Contoh data seed 2",
	},
	{
		JenisProduct: "Monitor", SerialNumber: "SN-00125", Name: "RTI-GAMMA-003", Proyek: "GAMMA",
		Status: "Broken", Lokasi: "Jakarta HQ", Deskripsi: "Contoh data seed 3",
	},
	{
		JenisProduct: "Server", SerialNumber: "SN-00126", Name: "RTI-ALPHA-004", Proyek: "ALPHA",
		PasswordPin: "9999", Account: "svcacct", PasswordAccount: "secret04", IPAddress: "10.0.0.14",
		Anydesk: "345 678 901", Rustdesk: "rd-004", PasswordAnydeskRustdesk: "rdpass04",
		LicenseWindows: "Server 2022", Status: "Healthy", Lokasi: "Jakarta HQ",
		Deskripsi: "Contoh data seed 4",
	},
}

func main() {
	reset := flag.Bool("reset", false, "delete all existing items before seeding")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on already-set environment variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := db.Connect(ctx, db.LoadConfig())
	cancel()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Disconnect(disconnectCtx, client)
	}()

	if *reset {
		res, err := client.Database.Collection("items").DeleteMany(context.Background(), map[string]interface{}{})
		if err != nil {
			log.Fatalf("failed to clear items collection: %v", err)
		}
		log.Printf("cleared %d existing item(s)\n", res.DeletedCount)
	}

	repo := repository.NewItemRepository(client.Database)
	for _, item := range seedItems {
		created, err := repo.Create(context.Background(), item)
		if err != nil {
			log.Fatalf("failed to seed item %q: %v", item.Name, err)
		}
		log.Printf("seeded %s (_id=%s)\n", created.Name, created.ID)
	}

	log.Printf("done: seeded %d item(s)\n", len(seedItems))
}
