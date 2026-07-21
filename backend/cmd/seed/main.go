// Command seed populates MongoDB with a handful of sample projects and items
// for local development and manual testing.
//
// Usage:
//
//	cd backend
//	go run ./cmd/seed          # insert the sample data (additive)
//	go run ./cmd/seed --reset  # delete all existing items/projects first, then insert
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

// seedProjects mirrors the Proyek entity: id (assigned on insert), namaProyek, lokasi.
var seedProjects = []models.Project{
	{NamaProyek: "ALPHA", Lokasi: "Jakarta HQ"},
	{NamaProyek: "BETA", Lokasi: "Surabaya Branch"},
	{NamaProyek: "GAMMA", Lokasi: "Jakarta HQ"},
}

// seedItemsFor builds sample items referencing the given project IDs (keyed
// by namaProyek). Status values match the frontend's canonical set:
// Healthy | Under Maintenance | Broken.
func seedItemsFor(projectIDs map[string]string) []models.Item {
	return []models.Item{
		{
			Jenis: "Laptop", SerialNumber: "SN-00123", Nama: "RTI-ALPHA-001", IdProyek: projectIDs["ALPHA"],
			Credentials: models.Credentials{"account": "user01", "passwordAccount": "secret01", "passwordPin": "1234"},
			RemoteInfo: models.RemoteInfo{
				"ipAddress": "10.0.0.12", "anydesk": "123 456 789", "passwordRemote": "rdpass01",
			},
			LicenseWindows: "Pro", LicenseOffice: "365", Status: "Healthy",
			Deskripsi:        "Contoh data seed",
			CustomAttributes: map[string]interface{}{"warranty": "3 tahun"},
		},
		{
			Jenis: "PC", SerialNumber: "SN-00124", Nama: "RTI-BETA-002", IdProyek: projectIDs["BETA"],
			Credentials: models.Credentials{"account": "user02", "passwordAccount": "secret02", "passwordPin": "5678"},
			RemoteInfo: models.RemoteInfo{
				"ipAddress": "10.0.0.13", "rustdesk": "rd-002", "passwordRemote": "rdpass02",
			},
			LicenseWindows: "Home", LicenseOffice: "2021", Status: "Under Maintenance",
			Deskripsi: "Contoh data seed 2",
		},
		{
			Jenis: "Monitor", SerialNumber: "SN-00125", Nama: "RTI-GAMMA-003", IdProyek: projectIDs["GAMMA"],
			Status: "Broken", Deskripsi: "Contoh data seed 3",
		},
		{
			Jenis: "Server", SerialNumber: "SN-00126", Nama: "RTI-ALPHA-004", IdProyek: projectIDs["ALPHA"],
			Credentials: models.Credentials{"account": "svcacct", "passwordAccount": "secret04", "passwordPin": "9999"},
			RemoteInfo: models.RemoteInfo{
				"ipAddress": "10.0.0.14", "anydesk": "345 678 901", "rustdesk": "rd-004", "passwordRemote": "rdpass04",
			},
			LicenseWindows: "Server 2022", Status: "Healthy",
			Deskripsi: "Contoh data seed 4",
		},
	}
}

func main() {
	reset := flag.Bool("reset", false, "delete all existing items/projects before seeding")
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
		itemsRes, err := client.Database.Collection("items").DeleteMany(context.Background(), map[string]interface{}{})
		if err != nil {
			log.Fatalf("failed to clear items collection: %v", err)
		}
		log.Printf("cleared %d existing item(s)\n", itemsRes.DeletedCount)

		projectsRes, err := client.Database.Collection("projects").DeleteMany(context.Background(), map[string]interface{}{})
		if err != nil {
			log.Fatalf("failed to clear projects collection: %v", err)
		}
		log.Printf("cleared %d existing project(s)\n", projectsRes.DeletedCount)
	}

	projectRepo := repository.NewProjectRepository(client.Database)
	projectIDs := make(map[string]string, len(seedProjects))
	for _, project := range seedProjects {
		created, err := projectRepo.Create(context.Background(), project)
		if err != nil {
			log.Fatalf("failed to seed project %q: %v", project.NamaProyek, err)
		}
		projectIDs[created.NamaProyek] = created.ID
		log.Printf("seeded project %s (_id=%s)\n", created.NamaProyek, created.ID)
	}

	itemRepo := repository.NewItemRepository(client.Database)
	items := seedItemsFor(projectIDs)
	for _, item := range items {
		created, err := itemRepo.Create(context.Background(), item)
		if err != nil {
			log.Fatalf("failed to seed item %q: %v", item.Nama, err)
		}
		log.Printf("seeded item %s (_id=%s)\n", created.Nama, created.ID)
	}

	log.Printf("done: seeded %d project(s), %d item(s)\n", len(seedProjects), len(items))
}
