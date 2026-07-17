package models

type Project struct {
	ID         string `bson:"_id" json:"_id"`
	NamaProyek string `bson:"namaProyek" json:"namaProyek"`
	Lokasi     string `bson:"lokasi" json:"lokasi"`
}
