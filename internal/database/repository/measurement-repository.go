package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	//	database "github.com/cristiandpt/healthcare/measures-consumer/internal/database"      // Replace with your actual module path
	entity "github.com/cristiandpt/healthcare/measures-consumer/internal/database/entity" // Replace with your actual module path
)

const measurementsCollection = "measurements"

// UserRepository handles database operations for users.
type MeasurementRepository struct {
	db *mongo.Database
}

// NewUserRepository creates a new UserRepository instance.
func NewMeasurementRepository(db *mongo.Database) *MeasurementRepository {
	return &MeasurementRepository{db: db}
}

// CreateUser inserts a new user into the database.
func (r *MeasurementRepository) CreateMeasurementRepository(ctx context.Context, user *entity.Measurements) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.Collection(measurementsCollection).InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByID retrieves a user by their ID.
func (r *MeasurementRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*entity.Measurements, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var measurements entity.Measurements
	err := r.db.Collection(measurementsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&measurements)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &measurements, nil
}

// GetAllUsers retrieves all users from the database.
func (r *MeasurementRepository) GetAllUsers(ctx context.Context) ([]*entity.Measurements, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.db.Collection(measurementsCollection).Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*entity.Measurements
	for cursor.Next(ctx) {
		var user entity.Measurements
		if err := cursor.Decode(&user); err != nil {
			return nil, fmt.Errorf("failed to decode user: %w", err)
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return users, nil
}

// DeleteUser deletes a user by their ID.
func (r *MeasurementRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.Collection(measurementsCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
