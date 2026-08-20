// Command seed populates MongoDB with sample projects, item types and a
// varied set of ~120 items for local development and manual testing.
//
// Usage:
//
//	cd backend
//	go run ./cmd/seed          # insert the sample data (additive)
//	go run ./cmd/seed --reset  # delete all existing items/projects first, then insert
//
// It connects using the same MONGODB_URI / MONGODB_DB as cmd/server (loaded
// from backend/.env if present), so make sure that's set up first.
//
// The item set is generated, not hand-written, but it is *deterministic*: the
// PRNG is seeded with a fixed constant, so two runs of `--reset` produce the
// same rows (bar the `_id`s, the "now" the timestamps are measured back from,
// and the key order inside the schemaless credentials/remoteInfo/
// customAttributes maps, which follows Go's randomized map iteration).
// That keeps manual testing repeatable while still giving the
// dashboard enough spread — every jenis, every status, items
// with and without credentials/remoteInfo/licenses/customAttributes, and
// createdAt/updatedAt spread over the last ~18 months so date-range filters
// and sorting have something to chew on.
//
// The project set also covers the states the map and the delete rules care
// about, each of which is otherwise easy to leave untested:
//
//	mapped + items    ALPHA, BETA, GAMMA, DELTA, EPSILON, ZETA  -> ordinary pins
//	unmapped + items  ETA, THETA                                -> /analytics/map "unmapped"
//	mapped + empty    IOTA                                      -> pin with totalItems 0,
//	                                                               and the only project whose
//	                                                               DELETE actually succeeds
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/models"
	"my-backend/internal/repository"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Fixed PRNG seed — see the package comment for why this is deterministic.
const (
	prngSeed1 = 20260805
	prngSeed2 = 0x5eed
)

// seedProjects mirrors the Proyek entity: id (assigned on insert), namaProyek,
// lokasi, koordinat.
//
// Coordinates are the real city centres, spread across Indonesia so a map
// render has something with actual extent to fit bounds to. ETA and THETA are
// left without coordinates on purpose: "project exists but isn't mapped yet"
// is a state the UI has to handle (it shows up under `unmapped` in
// /analytics/map), and it can't be tested against a dataset where every
// project happens to have a point.
func koordinat(lat, lng float64) *models.Koordinat {
	return &models.Koordinat{Lat: lat, Lng: lng}
}

var seedProjects = []models.Project{
	{NamaProyek: "ALPHA", Lokasi: "Jakarta HQ", Koordinat: koordinat(-6.2088, 106.8456)},
	{NamaProyek: "BETA", Lokasi: "Surabaya Branch", Koordinat: koordinat(-7.2575, 112.7521)},
	{NamaProyek: "GAMMA", Lokasi: "Bandung Site", Koordinat: koordinat(-6.9175, 107.6191)},
	{NamaProyek: "DELTA", Lokasi: "Medan Warehouse", Koordinat: koordinat(3.5952, 98.6722)},
	{NamaProyek: "EPSILON", Lokasi: "Makassar Site", Koordinat: koordinat(-5.1477, 119.4327)},
	{NamaProyek: "ZETA", Lokasi: "Balikpapan Field Office", Koordinat: koordinat(-1.2379, 116.8529)},
	{NamaProyek: "ETA", Lokasi: "Semarang Plant"},
	{NamaProyek: "THETA", Lokasi: "Denpasar Office"},
	{NamaProyek: "IOTA", Lokasi: "Yogyakarta Office", Koordinat: koordinat(-7.7956, 110.3695)},
}

// projectWeights decides how many items land on each project. HQ and the two
// big branches carry most of the fleet, the small sites carry a handful —
// an evenly spread set would make the project filter look artificial.
//
// IOTA is weight 0 by design: a project with no items at all. Without one,
// three behaviors have no data to exercise — DELETE /projects/{id} never
// succeeds (every other project is protected by its items), the
// projectsWithoutItems counter in /analytics/summary is always 0, and
// /analytics/map never returns a pin with an empty byStatus.
var projectWeights = map[string]int{
	"ALPHA": 5, "BETA": 3, "GAMMA": 3, "DELTA": 2,
	"EPSILON": 2, "ZETA": 1, "ETA": 2, "THETA": 1,
	"IOTA": 0,
}

// statusWeights matches the frontend's canonical status set
// (Healthy | Under Maintenance | Broken), skewed the way a real fleet is.
var statusWeights = []struct {
	status string
	weight int
}{
	{"Healthy", 7},
	{"Under Maintenance", 2},
	{"Broken", 1},
}

// jenisSpec describes one item type and how its items are generated. Every
// jenis listed here is also registered in the itemTypes master collection, so
// the invariant "every jenis in use appears in the master list" holds for a
// freshly seeded database without relying on the write-time auto-register.
type jenisSpec struct {
	jenis        string
	count        int
	serialPrefix string
	// models are the concrete hardware models picked from at random; the
	// chosen one goes into deskripsi and (where relevant) customAttributes.
	models []string
	// notes are deskripsi suffixes — what the unit is actually used for.
	notes []string
	// windows/office licenses; empty means this jenis has no such license
	// (a monitor or a switch does not run Windows), which is exactly the
	// "field legitimately blank" case the UI needs to render.
	windows []string
	office  []string
	// credRate/remoteRate are the 0..10 odds that an item of this jenis
	// carries credentials / remoteInfo at all.
	credRate   int
	remoteRate int
	// attrs builds the customAttributes for one item, or returns nil.
	attrs func(rng *rand.Rand, model string) map[string]interface{}
}

var (
	windowsClient = []string{"Pro", "Home", "Pro 11", "Enterprise", ""}
	officeSuites  = []string{"365", "2021", "2019", "LibreOffice", ""}
	vendors       = []string{"PT Sinar Teknologi", "CV Mitra Komputer", "PT Datacomm Nusantara", "Toko Anugerah IT"}
)

func pickAttrs(rng *rand.Rand, pairs map[string]interface{}, chance int) map[string]interface{} {
	if rng.IntN(10) >= chance {
		return nil
	}
	return pairs
}

var seedJenis = []jenisSpec{
	{
		jenis: "Laptop", count: 24, serialPrefix: "LTP",
		models:  []string{"Lenovo ThinkPad T14", "Dell Latitude 5430", "HP EliteBook 840", "Asus ExpertBook B9", "Acer TravelMate P4"},
		notes:   []string{"unit staf operasional", "unit engineer lapangan", "unit admin proyek", "unit cadangan tim IT"},
		windows: windowsClient, office: officeSuites,
		credRate: 9, remoteRate: 7,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"ram":           []string{"8 GB", "16 GB", "32 GB"}[rng.IntN(3)],
				"storage":       []string{"256 GB SSD", "512 GB SSD", "1 TB SSD"}[rng.IntN(3)],
				"garansiTahun":  rng.IntN(3) + 1,
				"tanggalBeli":   fmt.Sprintf("202%d-%02d-%02d", rng.IntN(5)+1, rng.IntN(12)+1, rng.IntN(28)+1),
				"vendorPembeli": vendors[rng.IntN(len(vendors))],
			}, 8)
		},
	},
	{
		jenis: "PC", count: 20, serialPrefix: "PCD",
		models:  []string{"HP ProDesk 400 G9", "Dell OptiPlex 7010", "Lenovo ThinkCentre M70q", "Rakitan Ryzen 5 5600G"},
		notes:   []string{"workstation ruang admin", "workstation front office", "unit ruang meeting", "unit operator gudang"},
		windows: windowsClient, office: officeSuites,
		credRate: 8, remoteRate: 8,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"cpu":          []string{"Intel i5-12500", "Intel i7-12700", "Ryzen 5 5600G", "Ryzen 7 5700G"}[rng.IntN(4)],
				"ram":          []string{"8 GB", "16 GB"}[rng.IntN(2)],
				"garansiTahun": rng.IntN(4),
			}, 7)
		},
	},
	{
		jenis: "Monitor", count: 16, serialPrefix: "MON",
		models:   []string{"LG 24MK600", "Dell P2422H", "Samsung LF24T350", "AOC 24B2XH", "Philips 243V7"},
		notes:    []string{"pendamping workstation admin", "layar kedua tim engineering", "layar ruang rapat", "layar operator"},
		credRate: 0, remoteRate: 0,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"ukuranInci": []interface{}{21.5, 24, 27}[rng.IntN(3)],
				"resolusi":   []string{"1920x1080", "2560x1440"}[rng.IntN(2)],
			}, 6)
		},
	},
	{
		jenis: "Server", count: 8, serialPrefix: "SRV",
		models:   []string{"Dell PowerEdge R450", "HPE ProLiant DL360 Gen10", "Lenovo ThinkSystem SR630", "Supermicro SYS-1029P"},
		notes:    []string{"file server proyek", "aplikasi internal", "database node", "backup node"},
		windows:  []string{"Server 2019", "Server 2022", ""},
		credRate: 10, remoteRate: 10,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"rackUnit":     fmt.Sprintf("RU-%02d", rng.IntN(42)+1),
				"ram":          []string{"32 GB", "64 GB", "128 GB"}[rng.IntN(3)],
				"raid":         []string{"RAID 1", "RAID 5", "RAID 10"}[rng.IntN(3)],
				"garansiTahun": rng.IntN(3) + 3,
			}, 9)
		},
	},
	{
		jenis: "Printer", count: 8, serialPrefix: "PRN",
		models:   []string{"Epson L3210", "Canon LBP2900", "Brother HL-L2375DW", "HP LaserJet M404dn"},
		notes:    []string{"printer ruang admin", "printer gudang", "printer front office"},
		credRate: 2, remoteRate: 5,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"tipeTinta": []string{"Ink Tank", "Toner"}[rng.IntN(2)],
				"jaringan":  rng.IntN(2) == 0,
			}, 6)
		},
	},
	{
		jenis: "Scanner", count: 3, serialPrefix: "SCN",
		models:   []string{"Epson DS-530", "Canon DR-C225", "Fujitsu ScanSnap iX1400"},
		notes:    []string{"digitalisasi dokumen proyek", "arsip administrasi"},
		credRate: 0, remoteRate: 0,
	},
	{
		jenis: "Router", count: 6, serialPrefix: "RTR",
		models:   []string{"MikroTik RB4011", "MikroTik hEX S", "Cisco ISR 1111", "TP-Link ER605"},
		notes:    []string{"gateway utama kantor", "gateway cadangan", "router WAN site"},
		credRate: 9, remoteRate: 9,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"firmware": fmt.Sprintf("v%d.%d", rng.IntN(4)+6, rng.IntN(10)),
				"vlan":     rng.IntN(2) == 0,
			}, 7)
		},
	},
	{
		jenis: "Switch", count: 6, serialPrefix: "SWT",
		models:   []string{"Cisco CBS250-24T", "TP-Link TL-SG1024", "MikroTik CRS326", "Ubiquiti USW-24"},
		notes:    []string{"distribusi lantai 1", "distribusi lantai 2", "switch rack utama"},
		credRate: 7, remoteRate: 8,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"jumlahPort": []interface{}{8, 16, 24, 48}[rng.IntN(4)],
				"managed":    rng.IntN(3) > 0,
			}, 7)
		},
	},
	{
		jenis: "Access Point", count: 6, serialPrefix: "WAP",
		models:   []string{"Ubiquiti U6-Lite", "TP-Link EAP245", "Ruijie RG-AP720", "MikroTik cAP ax"},
		notes:    []string{"wifi area kerja", "wifi ruang rapat", "wifi area gudang"},
		credRate: 6, remoteRate: 7,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"ssid":    []string{"RTI-Staff", "RTI-Guest", "RTI-Ops"}[rng.IntN(3)],
				"band":    []string{"2.4/5 GHz", "5 GHz"}[rng.IntN(2)],
				"poePort": true,
			}, 6)
		},
	},
	{
		jenis: "Firewall", count: 3, serialPrefix: "FWL",
		models:   []string{"FortiGate 60F", "Sophos XGS 116", "pfSense Netgate 4100"},
		notes:    []string{"perimeter kantor pusat", "perimeter cabang"},
		credRate: 10, remoteRate: 10,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return map[string]interface{}{
				"lisensiUTMBerakhir": fmt.Sprintf("202%d-%02d-01", rng.IntN(3)+6, rng.IntN(12)+1),
				"haPair":             rng.IntN(2) == 0,
			}
		},
	},
	{
		jenis: "UPS", count: 5, serialPrefix: "UPS",
		models:   []string{"APC BX1100C", "ICA CE-1200", "Prolink PRO700SFC", "Eaton 5E 1500"},
		notes:    []string{"backup daya rack server", "backup daya workstation admin", "backup daya CCTV"},
		credRate: 0, remoteRate: 3,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"kapasitasVA":     []interface{}{650, 1100, 1500, 3000}[rng.IntN(4)],
				"tahunGantiBatre": 2024 + rng.IntN(3),
			}, 8)
		},
	},
	{
		jenis: "Proyektor", count: 4, serialPrefix: "PRJ",
		models:   []string{"Epson EB-X06", "BenQ MS550", "Acer X1226AH"},
		notes:    []string{"ruang rapat utama", "ruang training", "unit presentasi keliling"},
		credRate: 0, remoteRate: 0,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"lumens":   []interface{}{3300, 3600, 4000}[rng.IntN(3)],
				"jamLampu": rng.IntN(4000),
			}, 7)
		},
	},
	{
		jenis: "CCTV", count: 6, serialPrefix: "CAM",
		models:   []string{"Hikvision DS-2CD1043", "Dahua IPC-HFW1230", "Ezviz C3W", "Uniview IPC2122"},
		notes:    []string{"pengawasan pintu masuk", "pengawasan area gudang", "pengawasan ruang server"},
		credRate: 8, remoteRate: 9,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"resolusi":    []string{"2 MP", "4 MP", "5 MP"}[rng.IntN(3)],
				"penyimpanan": []string{"NVR 1 TB", "NVR 2 TB", "MicroSD 128 GB"}[rng.IntN(3)],
			}, 7)
		},
	},
	{
		jenis: "NAS", count: 3, serialPrefix: "NAS",
		models:   []string{"Synology DS220+", "QNAP TS-464", "Synology DS923+"},
		notes:    []string{"arsip dokumen proyek", "target backup harian"},
		credRate: 10, remoteRate: 9,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return map[string]interface{}{
				"kapasitasTB": []interface{}{4, 8, 16}[rng.IntN(3)],
				"raid":        []string{"RAID 1", "SHR"}[rng.IntN(2)],
			}
		},
	},
	{
		jenis: "Tablet", count: 4, serialPrefix: "TAB",
		models:   []string{"Samsung Galaxy Tab A8", "iPad 9th Gen", "Lenovo Tab M10"},
		notes:    []string{"checklist inspeksi lapangan", "presentasi klien", "unit survei"},
		credRate: 5, remoteRate: 2,
		attrs: func(rng *rand.Rand, model string) map[string]interface{} {
			return pickAttrs(rng, map[string]interface{}{
				"os":      []string{"Android 13", "Android 14", "iPadOS 17"}[rng.IntN(3)],
				"simCard": rng.IntN(2) == 0,
			}, 6)
		},
	},
	{
		// "Lainnya" is just an ordinary catch-all type — it carries no special
		// meaning in the API (it used to be the reassign target when a type was
		// deleted; deletes are now blocked instead). Kept in the seed set
		// because a real inventory always has a miscellaneous bucket.
		jenis: "Lainnya", count: 4, serialPrefix: "MSC",
		models:   []string{"Barcode Scanner Zebra DS2208", "Label Printer Brother QL-800", "Docking Station Dell WD19", "Headset Jabra Evolve 20"},
		notes:    []string{"aksesori pendukung operasional", "peralatan gudang", "belum dikategorikan"},
		credRate: 0, remoteRate: 0,
	},
}

// seedItemTypes is derived from seedJenis so the master list can never drift
// from the jenis actually in use.
func seedItemTypes() []models.ItemType {
	types := make([]models.ItemType, 0, len(seedJenis))
	for _, spec := range seedJenis {
		types = append(types, models.ItemType{Jenis: spec.jenis})
	}
	return types
}

func pick[T any](rng *rand.Rand, options []T) T {
	return options[rng.IntN(len(options))]
}

// pickOptional returns "" for an empty option set, so a jenis with no Windows
// or Office license leaves those fields blank rather than inventing one.
func pickOptional(rng *rand.Rand, options []string) string {
	if len(options) == 0 {
		return ""
	}
	return pick(rng, options)
}

func weightedProjects() []string {
	// Built from seedProjects (not by ranging the map) so the order — and
	// therefore every draw made from it — stays deterministic.
	weighted := make([]string, 0, 32)
	for _, project := range seedProjects {
		for i := 0; i < projectWeights[project.NamaProyek]; i++ {
			weighted = append(weighted, project.NamaProyek)
		}
	}
	return weighted
}

func pickStatus(rng *rand.Rand) string {
	total := 0
	for _, entry := range statusWeights {
		total += entry.weight
	}
	roll := rng.IntN(total)
	for _, entry := range statusWeights {
		if roll < entry.weight {
			return entry.status
		}
		roll -= entry.weight
	}
	return statusWeights[0].status
}

const isoLayout = "2006-01-02T15:04:05.000Z"

// timestampsFor spreads createdAt over the last ~18 months and puts updatedAt
// somewhere between createdAt and now, so sorting and the createdFrom/
// createdTo / updatedFrom/updatedTo range filters have real spread to work
// with. Items are written straight to the collection (not through
// ItemRepository.Create) precisely because the repository stamps both fields
// with "now" — which would leave every seeded row indistinguishable by date.
func timestampsFor(rng *rand.Rand, now time.Time) (string, string) {
	createdAgo := time.Duration(rng.IntN(540))*24*time.Hour + time.Duration(rng.IntN(24))*time.Hour + time.Duration(rng.IntN(60))*time.Minute
	created := now.Add(-createdAgo)
	updated := created.Add(time.Duration(rng.Int64N(int64(now.Sub(created)) + 1)))
	return created.UTC().Format(isoLayout), updated.UTC().Format(isoLayout)
}

func credentialsFor(rng *rand.Rand, spec jenisSpec, seq int) models.Credentials {
	if rng.IntN(10) >= spec.credRate {
		return nil
	}
	creds := models.Credentials{
		"account":         fmt.Sprintf("%s%03d", []string{"user", "admin", "ops", "svc"}[rng.IntN(4)], seq),
		"passwordAccount": fmt.Sprintf("Pw-%s-%04d", spec.serialPrefix, rng.IntN(10000)),
	}
	if rng.IntN(2) == 0 {
		creds["passwordPin"] = fmt.Sprintf("%04d", rng.IntN(10000))
	}
	// A few items carry ad-hoc keys under credentials — the schemaless shape
	// is a deliberate feature, and the seed should exercise it.
	if rng.IntN(8) == 0 {
		creds["domain"] = "RTI-LOCAL"
	}
	if rng.IntN(12) == 0 {
		creds["apiToken"] = fmt.Sprintf("tok_%08x", rng.Uint32())
	}
	return creds
}

func remoteInfoFor(rng *rand.Rand, spec jenisSpec, seq int) models.RemoteInfo {
	if rng.IntN(10) >= spec.remoteRate {
		return nil
	}
	ip := fmt.Sprintf("10.%d.%d.%d", rng.IntN(4)+1, rng.IntN(20), seq%250+2)
	remote := models.RemoteInfo{"ipAddress": ip}
	switch rng.IntN(4) {
	case 0:
		remote["anydesk"] = fmt.Sprintf("%03d %03d %03d", rng.IntN(1000), rng.IntN(1000), rng.IntN(1000))
	case 1:
		remote["rustdesk"] = fmt.Sprintf("rd-%09d", rng.IntN(1000000000))
	case 2:
		remote["anydesk"] = fmt.Sprintf("%03d %03d %03d", rng.IntN(1000), rng.IntN(1000), rng.IntN(1000))
		remote["rustdesk"] = fmt.Sprintf("rd-%09d", rng.IntN(1000000000))
	}
	if rng.IntN(3) > 0 {
		remote["passwordRemote"] = fmt.Sprintf("rm-%s-%04d", spec.serialPrefix, rng.IntN(10000))
	}
	// A management web UI only makes sense on gear that has one, and it must
	// point at the same address as ipAddress rather than a second random one.
	if spec.remoteRate >= 8 && rng.IntN(3) == 0 {
		remote["webUi"] = "https://" + ip
	}
	return remote
}

// seedItemsFor builds the sample items referencing the given project IDs
// (keyed by namaProyek).
func seedItemsFor(projectIDs map[string]string, now time.Time) []models.Item {
	rng := rand.New(rand.NewPCG(prngSeed1, prngSeed2))
	weighted := weightedProjects()

	perProject := make(map[string]int, len(projectIDs))
	items := make([]models.Item, 0, 128)
	seq := 0

	for _, spec := range seedJenis {
		for i := 0; i < spec.count; i++ {
			seq++
			projectName := pick(rng, weighted)
			perProject[projectName]++
			model := pick(rng, spec.models)

			created, updated := timestampsFor(rng, now)

			item := models.Item{
				Jenis:          spec.jenis,
				SerialNumber:   fmt.Sprintf("%s-%05d", spec.serialPrefix, 10000+seq*7),
				Nama:           fmt.Sprintf("RTI-%s-%03d", projectName, perProject[projectName]),
				IdProyek:       projectIDs[projectName],
				Credentials:    credentialsFor(rng, spec, seq),
				RemoteInfo:     remoteInfoFor(rng, spec, seq),
				LicenseWindows: pickOptional(rng, spec.windows),
				LicenseOffice:  pickOptional(rng, spec.office),
				Status:         pickStatus(rng),
				Deskripsi:      fmt.Sprintf("%s — %s", model, pick(rng, spec.notes)),
				CreatedAt:      created,
				UpdatedAt:      updated,
			}
			if spec.attrs != nil {
				item.CustomAttributes = spec.attrs(rng, model)
			}
			// nama is optional in the API; a couple of rows leave it blank so
			// the UI's "unnamed item" path actually gets exercised by the
			// seed instead of only in tests.
			if seq%61 == 0 {
				item.Nama = ""
			}
			items = append(items, item)
		}
	}
	return items
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

		itemTypesRes, err := client.Database.Collection("itemTypes").DeleteMany(context.Background(), map[string]interface{}{})
		if err != nil {
			log.Fatalf("failed to clear itemTypes collection: %v", err)
		}
		log.Printf("cleared %d existing item type(s)\n", itemTypesRes.DeletedCount)
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

	itemTypeRepo := repository.NewItemTypeRepository(client.Database)
	itemTypes := seedItemTypes()
	for _, itemType := range itemTypes {
		created, err := itemTypeRepo.Create(context.Background(), itemType)
		if err != nil {
			log.Fatalf("failed to seed item type %q: %v", itemType.Jenis, err)
		}
		log.Printf("seeded item type %s (_id=%s)\n", created.Jenis, created.ID)
	}

	items := seedItemsFor(projectIDs, time.Now())
	docs := make([]interface{}, 0, len(items))
	perJenis := make(map[string]int, len(seedJenis))
	for _, item := range items {
		item.ID = bson.NewObjectID().Hex()
		perJenis[item.Jenis]++
		docs = append(docs, item)
	}
	if _, err := client.Database.Collection("items").InsertMany(context.Background(), docs); err != nil {
		log.Fatalf("failed to seed items: %v", err)
	}
	for _, spec := range seedJenis {
		log.Printf("seeded %3d item(s) of jenis %s\n", perJenis[spec.jenis], spec.jenis)
	}

	// Spell out the project coverage, so it's obvious from the seed output
	// alone which projects are pins, which aren't mapped yet, and which one is
	// empty (and therefore the only one you can actually delete).
	var mapped, unmapped, empty []string
	for _, project := range seedProjects {
		if projectWeights[project.NamaProyek] == 0 {
			empty = append(empty, project.NamaProyek)
		}
		if project.Koordinat == nil {
			unmapped = append(unmapped, project.NamaProyek)
			continue
		}
		mapped = append(mapped, project.NamaProyek)
	}
	log.Printf("projects with koordinat (map pins): %v\n", mapped)
	log.Printf("projects without koordinat (unmapped): %v\n", unmapped)
	log.Printf("projects without items (deletable): %v\n", empty)

	log.Printf("done: seeded %d project(s), %d item(s), %d item type(s)\n", len(seedProjects), len(items), len(itemTypes))
}
