package models

// Koordinat is a project's single map point.
//
// Field order is {lat, lng} on purpose: it matches what frontend map
// libraries take (L.marker([lat, lng]), google.maps.LatLng{lat, lng}) rather
// than GeoJSON's [longitude, latitude], which is the classic source of pins
// landing in the wrong hemisphere. Stored as two plain numbers, not a GeoJSON
// Point — there's no 2dsphere index and no radius search in this phase.
type Koordinat struct {
	Lat float64 `bson:"lat" json:"lat"`
	Lng float64 `bson:"lng" json:"lng"`
}

type Project struct {
	ID         string `bson:"_id" json:"_id"`
	NamaProyek string `bson:"namaProyek" json:"namaProyek"`
	Lokasi     string `bson:"lokasi" json:"lokasi"`
	// Koordinat is optional and all-or-nothing: a project either has both lat
	// and lng or no point at all. nil means "not mapped", and the field is
	// omitted from responses entirely rather than sent as null — so clients
	// test for presence instead of for a zero coordinate that would otherwise
	// plot as a real place in the Gulf of Guinea.
	Koordinat *Koordinat `bson:"koordinat,omitempty" json:"koordinat,omitempty"`
}
