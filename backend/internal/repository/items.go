package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"my-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("repository: item not found")

const itemsCollection = "items"

// allowed equality filter fields for ListItems. Password-bearing fields are
// deliberately excluded, consistent with them being excluded from ItemPublic.
var listFilterFields = []string{
	"jenis", "serialNumber", "nama", "idProyek", "status",
	"licenseWindows", "licenseOffice",
	"credentials.account", "remoteInfo.ipAddress", "remoteInfo.anydesk", "remoteInfo.rustdesk",
}

// allowed sort fields for ListItems.
var listSortFields = map[string]bool{
	"nama": true, "createdAt": true, "updatedAt": true,
	"status": true, "idProyek": true, "jenis": true,
}

type ItemRepository struct {
	coll *mongo.Collection
}

func NewItemRepository(db *mongo.Database) *ItemRepository {
	return &ItemRepository{coll: db.Collection(itemsCollection)}
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func newID() string {
	return bson.NewObjectID().Hex()
}

// Create inserts item after assigning a fresh ID and timestamps.
func (r *ItemRepository) Create(ctx context.Context, item models.Item) (models.Item, error) {
	item.ID = newID()
	now := nowISO()
	item.CreatedAt = now
	item.UpdatedAt = now

	if _, err := r.coll.InsertOne(ctx, item); err != nil {
		return models.Item{}, err
	}
	return item, nil
}

type ListParams struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
	Filters   map[string]string

	CreatedFrom string
	CreatedTo   string
	UpdatedFrom string
	UpdatedTo   string
}

type ListResult struct {
	Items      []models.Item
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

// List returns a paginated, filtered, sorted set of items.
func (r *ItemRepository) List(ctx context.Context, params ListParams) (ListResult, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	limit := params.Limit
	if limit < 1 {
		limit = 20
	}

	filter := bson.M{}
	for _, field := range listFilterFields {
		if v, ok := params.Filters[field]; ok && v != "" {
			filter[field] = v
		}
	}
	// customAttributes keys are arbitrary, so they can't be enumerated in
	// listFilterFields — trust any customAttributes.<key> filter the handler
	// already validated the prefix on.
	for field, v := range params.Filters {
		if strings.HasPrefix(field, "customAttributes.") && v != "" {
			filter[field] = v
		}
	}
	if q, ok := params.Filters["q"]; ok && q != "" {
		filter["$or"] = bson.A{
			bson.M{"nama": bson.M{"$regex": q, "$options": "i"}},
			bson.M{"serialNumber": bson.M{"$regex": q, "$options": "i"}},
		}
	}

	if params.CreatedFrom != "" || params.CreatedTo != "" {
		rangeFilter := bson.M{}
		if params.CreatedFrom != "" {
			rangeFilter["$gte"] = params.CreatedFrom
		}
		if params.CreatedTo != "" {
			rangeFilter["$lte"] = params.CreatedTo
		}
		filter["createdAt"] = rangeFilter
	}
	if params.UpdatedFrom != "" || params.UpdatedTo != "" {
		rangeFilter := bson.M{}
		if params.UpdatedFrom != "" {
			rangeFilter["$gte"] = params.UpdatedFrom
		}
		if params.UpdatedTo != "" {
			rangeFilter["$lte"] = params.UpdatedTo
		}
		filter["updatedAt"] = rangeFilter
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return ListResult{}, err
	}

	sortField := "createdAt"
	if listSortFields[params.SortBy] || strings.HasPrefix(params.SortBy, "customAttributes.") {
		sortField = params.SortBy
	}
	sortOrder := -1
	if params.SortOrder == "asc" {
		sortOrder = 1
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return ListResult{}, err
	}
	defer cursor.Close(ctx)

	items := make([]models.Item, 0, limit)
	if err := cursor.All(ctx, &items); err != nil {
		return ListResult{}, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetByID returns a single item, or ErrNotFound.
func (r *ItemRepository) GetByID(ctx context.Context, id string) (models.Item, error) {
	var item models.Item
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, err
	}
	return item, nil
}

// Update applies a partial patch to an existing item and returns the updated document.
func (r *ItemRepository) Update(ctx context.Context, id string, patch map[string]interface{}) (models.Item, error) {
	delete(patch, "_id")
	patch["updatedAt"] = nowISO()

	after := options.After
	var item models.Item
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": patch},
		options.FindOneAndUpdate().SetReturnDocument(after),
	).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, err
	}
	return item, nil
}

// Delete removes an item by ID, or returns ErrNotFound.
func (r *ItemRepository) Delete(ctx context.Context, id string) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Distinct returns the distinct, non-empty values stored for a single field.
func (r *ItemRepository) Distinct(ctx context.Context, field string) ([]string, error) {
	var values []string
	if err := r.coll.Distinct(ctx, field, bson.M{"$expr": bson.M{"$ne": bson.A{"$" + field, ""}}}).Decode(&values); err != nil {
		return nil, err
	}
	if values == nil {
		values = []string{}
	}
	return values, nil
}

// FilterOptions returns the distinct values for each of the given fields.
func (r *ItemRepository) FilterOptions(ctx context.Context, fields ...string) (map[string][]string, error) {
	result := make(map[string][]string, len(fields))
	for _, field := range fields {
		values, err := r.Distinct(ctx, field)
		if err != nil {
			return nil, err
		}
		result[field] = values
	}
	return result, nil
}

type CountBucket struct {
	ID    string `bson:"_id" json:"_id"`
	Count int    `bson:"count" json:"count"`
}

// CountBy runs a group-by-count aggregation on the given field.
func (r *ItemRepository) CountBy(ctx context.Context, field string) ([]CountBucket, error) {
	pipeline := bson.A{
		bson.M{"$group": bson.M{"_id": "$" + field, "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"count": -1}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	buckets := make([]CountBucket, 0)
	if err := cursor.All(ctx, &buckets); err != nil {
		return nil, err
	}
	return buckets, nil
}

// Count returns the total number of items in the collection.
func (r *ItemRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

// RecentlyAdded returns the most recently created items, limited to n.
func (r *ItemRepository) RecentlyAdded(ctx context.Context, n int) ([]models.Item, error) {
	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(n))

	cursor, err := r.coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]models.Item, 0, n)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
