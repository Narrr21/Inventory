package mocks

import "encoding/json"

type Item struct {
	ID                      string `json:"_id"`
	JenisProduct            string `json:"jenisProduct"`
	SerialNumber            string `json:"serialNumber"`
	Name                    string `json:"name"`
	Proyek                  string `json:"proyek"`
	PasswordPin             string `json:"passwordPin"`
	Account                 string `json:"account"`
	PasswordAccount         string `json:"passwordAccount"`
	IPAddress               string `json:"ipAddress"`
	Anydesk                 string `json:"anydesk"`
	Rustdesk                string `json:"rustdesk"`
	PasswordAnydeskRustdesk string `json:"passwordAnydeskRustdesk"`
	LicenseWindows          string `json:"licenseWindows"`
	LicenseOffice           string `json:"licenseOffice"`
	Status                  string `json:"status"`
	Lokasi                  string `json:"lokasi"`
	Deskripsi               string `json:"deskripsi"`
	CreatedAt               string `json:"createdAt"`
	UpdatedAt               string `json:"updatedAt"`
}

type ItemPublic struct {
	ID             string `json:"_id"`
	JenisProduct   string `json:"jenisProduct"`
	SerialNumber   string `json:"serialNumber"`
	Name           string `json:"name"`
	Proyek         string `json:"proyek"`
	Account        string `json:"account"`
	IPAddress      string `json:"ipAddress"`
	Anydesk        string `json:"anydesk"`
	Rustdesk       string `json:"rustdesk"`
	LicenseWindows string `json:"licenseWindows"`
	LicenseOffice  string `json:"licenseOffice"`
	Status         string `json:"status"`
	Lokasi         string `json:"lokasi"`
	Deskripsi      string `json:"deskripsi"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func (i Item) Public() ItemPublic {
	return ItemPublic{
		ID:             i.ID,
		JenisProduct:   i.JenisProduct,
		SerialNumber:   i.SerialNumber,
		Name:           i.Name,
		Proyek:         i.Proyek,
		Account:        i.Account,
		IPAddress:      i.IPAddress,
		Anydesk:        i.Anydesk,
		Rustdesk:       i.Rustdesk,
		LicenseWindows: i.LicenseWindows,
		LicenseOffice:  i.LicenseOffice,
		Status:         i.Status,
		Lokasi:         i.Lokasi,
		Deskripsi:      i.Deskripsi,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
	}
}

var Items = []Item{
	{
		ID: "665f1a1a1a1a1a1a1a1a1a1a", JenisProduct: "Laptop", SerialNumber: "SN-00123",
		Name: "RTI-ALPHA-001", Proyek: "ALPHA", PasswordPin: "1234", Account: "user01",
		PasswordAccount: "secret01", IPAddress: "10.0.0.12", Anydesk: "123 456 789", Rustdesk: "",
		PasswordAnydeskRustdesk: "rdpass01", LicenseWindows: "Pro", LicenseOffice: "365",
		Status: "Active", Lokasi: "Jakarta HQ", Deskripsi: "Contoh data dummy",
		CreatedAt: "2026-07-01T02:00:00.000Z", UpdatedAt: "2026-07-01T02:00:00.000Z",
	},
	{
		ID: "665f1a1a1a1a1a1a1a1a1a1b", JenisProduct: "PC", SerialNumber: "SN-00124",
		Name: "RTI-BETA-002", Proyek: "BETA", PasswordPin: "5678", Account: "user02",
		PasswordAccount: "secret02", IPAddress: "10.0.0.13", Anydesk: "234 567 890", Rustdesk: "rd-002",
		PasswordAnydeskRustdesk: "rdpass02", LicenseWindows: "Home", LicenseOffice: "2021",
		Status: "Idle", Lokasi: "Surabaya Branch", Deskripsi: "Contoh data dummy 2",
		CreatedAt: "2026-07-01T03:00:00.000Z", UpdatedAt: "2026-07-01T03:00:00.000Z",
	},
	{
		ID: "665f1a1a1a1a1a1a1a1a1a1c", JenisProduct: "Monitor", SerialNumber: "SN-00125",
		Name: "RTI-GAMMA-003", Proyek: "GAMMA", PasswordPin: "", Account: "",
		PasswordAccount: "", IPAddress: "", Anydesk: "", Rustdesk: "",
		PasswordAnydeskRustdesk: "", LicenseWindows: "", LicenseOffice: "",
		Status: "Maintenance", Lokasi: "Jakarta HQ", Deskripsi: "Contoh data dummy 3",
		CreatedAt: "2026-07-02T02:00:00.000Z", UpdatedAt: "2026-07-02T02:00:00.000Z",
	},
	{
		ID: "665f1a1a1a1a1a1a1a1a1a1d", JenisProduct: "Server", SerialNumber: "SN-00126",
		Name: "RTI-ALPHA-004", Proyek: "ALPHA", PasswordPin: "9999", Account: "svcacct",
		PasswordAccount: "secret04", IPAddress: "10.0.0.14", Anydesk: "345 678 901", Rustdesk: "rd-004",
		PasswordAnydeskRustdesk: "rdpass04", LicenseWindows: "Server 2022", LicenseOffice: "",
		Status: "Active", Lokasi: "Jakarta HQ", Deskripsi: "Contoh data dummy 4",
		CreatedAt: "2026-07-02T04:00:00.000Z", UpdatedAt: "2026-07-02T04:00:00.000Z",
	},
}

// MergeItem overlays patch (raw decoded JSON body) onto base and returns the result,
// so handlers can echo back whatever the caller sent merged with mock data.
func MergeItem(base Item, patch map[string]interface{}) Item {
	baseBytes, _ := json.Marshal(base)
	var merged map[string]interface{}
	json.Unmarshal(baseBytes, &merged)

	for k, v := range patch {
		merged[k] = v
	}

	result := base
	mergedBytes, _ := json.Marshal(merged)
	json.Unmarshal(mergedBytes, &result)
	return result
}
