package models

import "encoding/json"

// Credentials and RemoteInfo are schemaless, like CustomAttributes — any
// top-level key sent under "credentials"/"remoteInfo" is preserved as-is,
// not limited to a fixed set of sub-fields. There is no server-side
// redaction of secret-looking keys (e.g. passwordPin): callers get back
// exactly what they stored, in both GetItem and ListItems.
type Credentials = map[string]interface{}

type RemoteInfo = map[string]interface{}

type Item struct {
	ID           string `bson:"_id" json:"_id"`
	Jenis        string `bson:"jenis" json:"jenis"`
	SerialNumber string `bson:"serialNumber" json:"serialNumber"`
	Nama         string `bson:"nama" json:"nama"`
	IdProyek     string `bson:"idProyek" json:"idProyek"`
	// NamaProyek is never stored — it's the referenced Project's own
	// namaProyek, resolved and set by the handler at response time, so it
	// can't go stale independently of the Project document and needs no
	// migration.
	NamaProyek       string                 `bson:"-" json:"namaProyek,omitempty"`
	Credentials      Credentials            `bson:"credentials" json:"credentials"`
	RemoteInfo       RemoteInfo             `bson:"remoteInfo" json:"remoteInfo"`
	LicenseWindows   string                 `bson:"licenseWindows" json:"licenseWindows"`
	LicenseOffice    string                 `bson:"licenseOffice" json:"licenseOffice"`
	Status           string                 `bson:"status" json:"status"`
	Deskripsi        string                 `bson:"deskripsi" json:"deskripsi"`
	CustomAttributes map[string]interface{} `bson:"customAttributes,omitempty" json:"customAttributes,omitempty"`
	CreatedAt        string                 `bson:"createdAt" json:"createdAt"`
	UpdatedAt        string                 `bson:"updatedAt" json:"updatedAt"`
}

type ItemPublic struct {
	ID               string                 `json:"_id"`
	Jenis            string                 `json:"jenis"`
	SerialNumber     string                 `json:"serialNumber"`
	Nama             string                 `json:"nama"`
	IdProyek         string                 `json:"idProyek"`
	NamaProyek       string                 `json:"namaProyek,omitempty"`
	Credentials      Credentials            `json:"credentials"`
	RemoteInfo       RemoteInfo             `json:"remoteInfo"`
	LicenseWindows   string                 `json:"licenseWindows"`
	LicenseOffice    string                 `json:"licenseOffice"`
	Status           string                 `json:"status"`
	Deskripsi        string                 `json:"deskripsi"`
	CustomAttributes map[string]interface{} `json:"customAttributes,omitempty"`
	CreatedAt        string                 `json:"createdAt"`
	UpdatedAt        string                 `json:"updatedAt"`
}

func (i Item) Public() ItemPublic {
	return ItemPublic{
		ID:               i.ID,
		Jenis:            i.Jenis,
		SerialNumber:     i.SerialNumber,
		Nama:             i.Nama,
		IdProyek:         i.IdProyek,
		NamaProyek:       i.NamaProyek,
		Credentials:      i.Credentials,
		RemoteInfo:       i.RemoteInfo,
		LicenseWindows:   i.LicenseWindows,
		LicenseOffice:    i.LicenseOffice,
		Status:           i.Status,
		Deskripsi:        i.Deskripsi,
		CustomAttributes: i.CustomAttributes,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}

// EnsureMaps normalizes nil Credentials/RemoteInfo to empty (non-nil) maps,
// so a response always shows "credentials": {} rather than "null" for an
// item that never had them set — nil maps marshal to JSON null, unlike the
// old fixed-shape structs which always had a zero value that encoded as {}.
func (i *Item) EnsureMaps() {
	if i.Credentials == nil {
		i.Credentials = Credentials{}
	}
	if i.RemoteInfo == nil {
		i.RemoteInfo = RemoteInfo{}
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
