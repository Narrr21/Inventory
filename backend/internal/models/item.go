package models

import "encoding/json"

type Credentials struct {
	Account         string `bson:"account,omitempty" json:"account,omitempty"`
	PasswordAccount string `bson:"passwordAccount,omitempty" json:"passwordAccount,omitempty"`
	PasswordPin     string `bson:"passwordPin,omitempty" json:"passwordPin,omitempty"`
}

type CredentialsPublic struct {
	Account string `bson:"account,omitempty" json:"account,omitempty"`
}

type RemoteInfo struct {
	IPAddress      string `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Anydesk        string `bson:"anydesk,omitempty" json:"anydesk,omitempty"`
	Rustdesk       string `bson:"rustdesk,omitempty" json:"rustdesk,omitempty"`
	PasswordRemote string `bson:"passwordRemote,omitempty" json:"passwordRemote,omitempty"`
}

type RemoteInfoPublic struct {
	IPAddress string `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Anydesk   string `bson:"anydesk,omitempty" json:"anydesk,omitempty"`
	Rustdesk  string `bson:"rustdesk,omitempty" json:"rustdesk,omitempty"`
}

type Item struct {
	ID               string                 `bson:"_id" json:"_id"`
	Jenis            string                 `bson:"jenis" json:"jenis"`
	SerialNumber     string                 `bson:"serialNumber" json:"serialNumber"`
	Nama             string                 `bson:"nama" json:"nama"`
	IdProyek         string                 `bson:"idProyek" json:"idProyek"`
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
	Credentials      CredentialsPublic      `json:"credentials"`
	RemoteInfo       RemoteInfoPublic       `json:"remoteInfo"`
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
		ID:           i.ID,
		Jenis:        i.Jenis,
		SerialNumber: i.SerialNumber,
		Nama:         i.Nama,
		IdProyek:     i.IdProyek,
		Credentials:  CredentialsPublic{Account: i.Credentials.Account},
		RemoteInfo: RemoteInfoPublic{
			IPAddress: i.RemoteInfo.IPAddress,
			Anydesk:   i.RemoteInfo.Anydesk,
			Rustdesk:  i.RemoteInfo.Rustdesk,
		},
		LicenseWindows:   i.LicenseWindows,
		LicenseOffice:    i.LicenseOffice,
		Status:           i.Status,
		Deskripsi:        i.Deskripsi,
		CustomAttributes: i.CustomAttributes,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
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
