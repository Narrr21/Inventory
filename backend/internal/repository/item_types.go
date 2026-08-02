package repository

import (
	"context"
	"regexp"
	"strings"

	"my-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const itemTypesCollection = "itemTypes"

type ItemTypeRepository struct {
	coll *mongo.Collection
}

func NewItemTypeRepository(db *mongo.Database) *ItemTypeRepository {
	return &ItemTypeRepository{coll: db.Collection(itemTypesCollection)}
}

// Create inserts an item type after assigning a fresh ID.
func (r *ItemTypeRepository) Create(ctx context.Context, itemType models.ItemType) (models.ItemType, error) {
	itemType.ID = newID()

	if _, err := r.coll.InsertOne(ctx, itemType); err != nil {
		return models.ItemType{}, err
	}
	return itemType, nil
}

// List returns all item types, sorted by jenis.
func (r *ItemTypeRepository) List(ctx context.Context) ([]models.ItemType, error) {
	findOpts := options.Find().SetSort(bson.D{{Key: "jenis", Value: 1}})

	cursor, err := r.coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	itemTypes := make([]models.ItemType, 0)
	if err := cursor.All(ctx, &itemTypes); err != nil {
		return nil, err
	}
	return itemTypes, nil
}

// ExistsByJenis reports whether an item type with the given jenis already
// exists, matched case-insensitively (unlike Project.NamaProyek, this field
// is free-typed by users, so "Laptop" and "laptop" are treated as the same
// entry to avoid suggestion-list clutter).
func (r *ItemTypeRepository) ExistsByJenis(ctx context.Context, jenis string) (bool, error) {
	pattern := "^" + regexp.QuoteMeta(strings.TrimSpace(jenis)) + "$"
	filter := bson.M{"jenis": bson.M{"$regex": pattern, "$options": "i"}}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Delete removes an item type by ID, or returns ErrNotFound. There is
// deliberately no check for referencing items — Item.Jenis is a free-text
// string, not a foreign key, so existing items are entirely unaffected by
// deleting a suggestion entry.
func (r *ItemTypeRepository) Delete(ctx context.Context, id string) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
