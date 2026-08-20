package repository

import (
	"context"
	"errors"

	"my-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const projectsCollection = "projects"

type ProjectRepository struct {
	coll *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) *ProjectRepository {
	return &ProjectRepository{coll: db.Collection(projectsCollection)}
}

// Create inserts a project after assigning a fresh ID.
func (r *ProjectRepository) Create(ctx context.Context, project models.Project) (models.Project, error) {
	project.ID = newID()

	if _, err := r.coll.InsertOne(ctx, project); err != nil {
		return models.Project{}, err
	}
	return project, nil
}

// List returns all projects, sorted by namaProyek.
func (r *ProjectRepository) List(ctx context.Context) ([]models.Project, error) {
	findOpts := options.Find().SetSort(bson.D{{Key: "namaProyek", Value: 1}})

	cursor, err := r.coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	projects := make([]models.Project, 0)
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetByID returns a single project, or ErrNotFound.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (models.Project, error) {
	var project models.Project
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Project{}, ErrNotFound
	}
	if err != nil {
		return models.Project{}, err
	}
	return project, nil
}

// GetByName returns a single project by exact (case-sensitive) namaProyek
// match, or ErrNotFound.
func (r *ProjectRepository) GetByName(ctx context.Context, namaProyek string) (models.Project, error) {
	var project models.Project
	err := r.coll.FindOne(ctx, bson.M{"namaProyek": namaProyek}).Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Project{}, ErrNotFound
	}
	if err != nil {
		return models.Project{}, err
	}
	return project, nil
}

// Exists reports whether a project with the given ID exists.
func (r *ProjectRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamaProyek reports whether a project with the given namaProyek
// (case-sensitive, exact match) already exists. If excludeID is non-empty,
// that project's own document is excluded from the check, so a PATCH can
// re-validate uniqueness without tripping on itself.
func (r *ProjectRepository) ExistsByNamaProyek(ctx context.Context, namaProyek string, excludeID string) (bool, error) {
	filter := bson.M{"namaProyek": namaProyek}
	if excludeID != "" {
		filter["_id"] = bson.M{"$ne": excludeID}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count returns the total number of projects.
func (r *ProjectRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

// Update replaces namaProyek/lokasi/koordinat on an existing project and
// returns the updated document, or ErrNotFound. A nil Koordinat is written as
// an $unset rather than a null field, so "project without a point" is always
// the same thing in the database (key absent) no matter whether the project
// never had one or had one removed.
func (r *ProjectRepository) Update(ctx context.Context, id string, project models.Project) (models.Project, error) {
	update := bson.M{
		"$set": bson.M{"namaProyek": project.NamaProyek, "lokasi": project.Lokasi},
	}
	if project.Koordinat != nil {
		update["$set"].(bson.M)["koordinat"] = project.Koordinat
	} else {
		update["$unset"] = bson.M{"koordinat": ""}
	}

	after := options.After
	var updated models.Project
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(after),
	).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Project{}, ErrNotFound
	}
	if err != nil {
		return models.Project{}, err
	}
	return updated, nil
}

// Delete removes a project by ID, or returns ErrNotFound. The "is any item
// still referencing this project?" check lives in the handler
// (handlers.ProjectHandler), which owns the 409 PROJECT_IN_USE response —
// keeping this a plain single-collection operation.
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
