package models

import "encoding/json"

type Item struct {
	ID                      string `bson:"_id" json:"_id"`
	JenisProduct            string `bson:"jenisProduct" json:"jenisProduct"`
	SerialNumber            string `bson:"serialNumber" json:"serialNumber"`
	Name                    string `bson:"name" json:"name"`
	Proyek                  string `bson:"proyek" json:"proyek"`
	PasswordPin             string `bson:"passwordPin" json:"passwordPin"`
	Account                 string `bson:"account" json:"account"`
	PasswordAccount         string `bson:"passwordAccount" json:"passwordAccount"`
	IPAddress               string `bson:"ipAddress" json:"ipAddress"`
	Anydesk                 string `bson:"anydesk" json:"anydesk"`
	Rustdesk                string `bson:"rustdesk" json:"rustdesk"`
	PasswordAnydeskRustdesk string `bson:"passwordAnydeskRustdesk" json:"passwordAnydeskRustdesk"`
	LicenseWindows          string `bson:"licenseWindows" json:"licenseWindows"`
	LicenseOffice           string `bson:"licenseOffice" json:"licenseOffice"`
	Status                  string `bson:"status" json:"status"`
	Lokasi                  string `bson:"lokasi" json:"lokasi"`
	Deskripsi               string `bson:"deskripsi" json:"deskripsi"`
	CreatedAt               string `bson:"createdAt" json:"createdAt"`
	UpdatedAt               string `bson:"updatedAt" json:"updatedAt"`
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

// MergeItem overlays patch (raw decoded JSON body) onto base and returns the result.
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
