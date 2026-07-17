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

// Exists reports whether a project with the given ID exists.
func (r *ProjectRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
