package models

// ItemType is a master-data entry for a "jenis barang" (item type) name —
// a lightweight, independently managed suggestion list for Item.Jenis. It is
// NOT a foreign key: Item.Jenis stays a free-text string regardless of
// whether a matching ItemType exists, so deleting one has no effect on
// existing items.
type ItemType struct {
	ID    string `bson:"_id" json:"_id"`
	Jenis string `bson:"jenis" json:"jenis"`
}
