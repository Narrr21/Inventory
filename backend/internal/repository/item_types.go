package repository

import (
	"context"
	"errors"
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

// caseInsensitiveExact builds a filter matching value exactly but ignoring
// case. Item type names are free-typed by users, so "Laptop" and "laptop"
// must resolve to the same master-list entry rather than two.
func caseInsensitiveExact(field, value string) bson.M {
	pattern := "^" + regexp.QuoteMeta(strings.TrimSpace(value)) + "$"
	return bson.M{field: bson.M{"$regex": pattern, "$options": "i"}}
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

// ListNames returns just the jenis names, sorted — the shape the
// filter-options dropdown consumes.
func (r *ItemTypeRepository) ListNames(ctx context.Context) ([]string, error) {
	itemTypes, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(itemTypes))
	for _, itemType := range itemTypes {
		names = append(names, itemType.Jenis)
	}
	return names, nil
}

// GetByID returns a single item type, or ErrNotFound.
func (r *ItemTypeRepository) GetByID(ctx context.Context, id string) (models.ItemType, error) {
	var itemType models.ItemType
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&itemType)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.ItemType{}, ErrNotFound
	}
	if err != nil {
		return models.ItemType{}, err
	}
	return itemType, nil
}

// ExistsByJenis reports whether an item type with the given jenis already
// exists, matched case-insensitively.
func (r *ItemTypeRepository) ExistsByJenis(ctx context.Context, jenis string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, caseInsensitiveExact("jenis", jenis))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// EnsureExists registers jenis in the master list if it isn't there yet, and
// is a no-op if it is (case-insensitively). This is what keeps the invariant
// "every Item.jenis in use also appears in the master list" true even when a
// client writes an item with a brand-new jenis without calling
// POST /item-types first. Blank input is ignored rather than treated as an
// error — required-field validation is the item handler's job, not this one's.
func (r *ItemTypeRepository) EnsureExists(ctx context.Context, jenis string) error {
	jenis = strings.TrimSpace(jenis)
	if jenis == "" {
		return nil
	}

	exists, err := r.ExistsByJenis(ctx, jenis)
	if err != nil || exists {
		return err
	}
	_, err = r.Create(ctx, models.ItemType{Jenis: jenis})
	return err
}

// Delete removes an item type by ID, or returns ErrNotFound. It deliberately
// does not check for referencing items — reassigning those items to the
// default type is the handler's job (see handlers.ItemTypeHandler), kept out
// of the repository so this stays a plain single-collection operation.
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
