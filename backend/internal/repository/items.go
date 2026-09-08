package repository

import (
	"context"
	"errors"
	"regexp"
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

// caseInsensitiveContains builds a substring match on field. The needle is
// regex-quoted because it comes straight from a search box: unescaped, "("
// makes Mongo reject the whole query (surfacing as a 500 on a perfectly
// reasonable search), and ".*" would match every document instead of the
// literal characters the user typed.
func caseInsensitiveContains(field, value string) bson.M {
	return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(value), "$options": "i"}}
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
	if limit > 100 {
		limit = 100
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
			caseInsensitiveContains("nama", q),
			caseInsensitiveContains("serialNumber", q),
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

// CountByJenis reports how many items currently use a jenis value, matched
// case-insensitively so it agrees with the case-insensitive uniqueness rule on
// the item-type master list ("Laptop" and "laptop" are the same type, so both
// block deleting it). Backs 409 ITEM_TYPE_IN_USE.
func (r *ItemRepository) CountByJenis(ctx context.Context, jenis string) (int64, error) {
	return r.coll.CountDocuments(ctx, caseInsensitiveExact("jenis", jenis))
}

// CountByProject reports how many items reference a project. Backs
// 409 PROJECT_IN_USE.
func (r *ItemRepository) CountByProject(ctx context.Context, idProyek string) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"idProyek": idProyek})
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

type CountBucket struct {
	ID    string `bson:"_id" json:"_id"`
	Count int    `bson:"count" json:"count"`
}

// CountBy runs a group-by-count aggregation on the given field, sorted by
// count desc then value asc so the order is stable across identical counts
// (an unstable order makes charts jump between refreshes for no reason).
//
// Items whose value is missing or blank are skipped: for licenseWindows/
// licenseOffice an empty string means "not applicable to this item", not a
// category worth a slice in a chart. Required fields (status, jenis) can't be
// blank anyway, so this only ever drops noise.
func (r *ItemRepository) CountBy(ctx context.Context, field string) ([]CountBucket, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{field: bson.M{"$nin": bson.A{"", nil}}}},
		bson.M{"$group": bson.M{"_id": "$" + field, "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.D{{Key: "count", Value: -1}, {Key: "_id", Value: 1}}},
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

// ProjectStatusBucket is one (project, status) pair and its item count.
type ProjectStatusBucket struct {
	IdProyek string `bson:"idProyek"`
	Status   string `bson:"status"`
	Count    int    `bson:"count"`
}

// CountByProjectAndStatus returns the item count for every (idProyek, status)
// combination in one aggregation — the raw material for both the per-project
// status breakdown on the map and the per-project totals in the summary.
// Doing it in a single query keeps /analytics/map at a fixed cost instead of
// one round trip per project.
func (r *ItemRepository) CountByProjectAndStatus(ctx context.Context) ([]ProjectStatusBucket, error) {
	pipeline := bson.A{
		bson.M{"$group": bson.M{
			"_id":   bson.M{"idProyek": "$idProyek", "status": "$status"},
			"count": bson.M{"$sum": 1},
		}},
		bson.M{"$project": bson.M{
			"_id":      0,
			"idProyek": "$_id.idProyek",
			"status":   "$_id.status",
			"count":    1,
		}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	buckets := make([]ProjectStatusBucket, 0)
	if err := cursor.All(ctx, &buckets); err != nil {
		return nil, err
	}
	return buckets, nil
}

// PeriodBucket is the number of items created in one YYYY-MM period.
type PeriodBucket struct {
	Period string `bson:"_id"`
	Count  int    `bson:"count"`
}

// CreatedPerMonth groups items by the YYYY-MM prefix of createdAt, ascending.
// createdAt is stored as an ISO-8601 UTC string, so the month is just its
// first 7 characters — no date parsing, and lexicographic order is also
// chronological order. Only months that actually have items come back; filling
// the empty months in between is the handler's job (it knows the window).
func (r *ItemRepository) CreatedPerMonth(ctx context.Context, fromPeriod string) ([]PeriodBucket, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"createdAt": bson.M{"$gte": fromPeriod}}},
		bson.M{"$group": bson.M{
			"_id":   bson.M{"$substrBytes": bson.A{"$createdAt", 0, 7}},
			"count": bson.M{"$sum": 1},
		}},
		bson.M{"$sort": bson.M{"_id": 1}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	buckets := make([]PeriodBucket, 0)
	if err := cursor.All(ctx, &buckets); err != nil {
		return nil, err
	}
	return buckets, nil
}

// CountCreatedBefore returns how many items were created strictly before an
// ISO-8601 timestamp (or YYYY-MM prefix). Used to seed the timeline's running
// total so the first bucket's cumulative includes everything older than the
// window rather than restarting at zero.
func (r *ItemRepository) CountCreatedBefore(ctx context.Context, iso string) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"createdAt": bson.M{"$lt": iso}})
}

// CountUpdatedBefore returns how many items haven't been touched since an
// ISO-8601 timestamp — the "stale, nobody has verified this in months" count.
func (r *ItemRepository) CountUpdatedBefore(ctx context.Context, iso string) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"updatedAt": bson.M{"$lt": iso}})
}

// CountOrphans returns how many items point at an idProyek that isn't in
// validProjectIDs. These can no longer be created (item writes validate
// idProyek, and deleting a project with items is blocked), so a non-zero
// result is leftover data from before that rule existed.
func (r *ItemRepository) CountOrphans(ctx context.Context, validProjectIDs []string) (int64, error) {
	ids := make(bson.A, 0, len(validProjectIDs))
	for _, id := range validProjectIDs {
		ids = append(ids, id)
	}
	return r.coll.CountDocuments(ctx, bson.M{"idProyek": bson.M{"$nin": ids}})
}

// Count returns the total number of items in the collection.
func (r *ItemRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

// RecentlyAdded returns the most recently created items, limited to n.
func (r *ItemRepository) RecentlyAdded(ctx context.Context, n int) ([]models.Item, error) {
	return r.recentBy(ctx, "createdAt", n)
}

// RecentlyUpdated returns the most recently modified items, limited to n.
func (r *ItemRepository) RecentlyUpdated(ctx context.Context, n int) ([]models.Item, error) {
	return r.recentBy(ctx, "updatedAt", n)
}

func (r *ItemRepository) recentBy(ctx context.Context, field string, n int) ([]models.Item, error) {
	findOpts := options.Find().
		SetSort(bson.D{{Key: field, Value: -1}}).
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
