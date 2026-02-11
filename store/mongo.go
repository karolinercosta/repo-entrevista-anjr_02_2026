package store

import (
	"context"
	"fmt"
	"time"

	"example.com/tasksapi/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
	col    *mongo.Collection
}

// NewMongo conecta no MongoDB e retorna a Store
func NewMongo(ctx context.Context, connectionString, dbName, collectionName string) (Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	db := client.Database(dbName)
	col := db.Collection(collectionName)

	return &MongoStore{
		client: client,
		db:     db,
		col:    col,
	}, nil
}

func (m *MongoStore) Create(t models.Task) models.Task {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = nil

	oid := primitive.NewObjectID()
	idHex := oid.Hex()
	t.ID = idHex

	// Leave DueDate as nil if not provided; convert zero time to nil
	if t.DueDate != nil && t.DueDate.IsZero() {
		t.DueDate = nil
	}

	doc := bson.M{
		"_id":         oid,
		"id":          idHex,
		"title":       t.Title,
		"description": t.Description,
		"status":      t.Status,
		"priority":    t.Priority,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
	}
	if t.DueDate != nil {
		doc["due_date"] = *t.DueDate
	}

	_, err := m.col.InsertOne(ctx, doc)
	if err != nil {
		return t
	}

	return t
}

func (m *MongoStore) List() []models.Task {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := m.col.Find(ctx, bson.M{})
	if err != nil {
		return []models.Task{}
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return []models.Task{}
	}

	return tasks
}

func (m *MongoStore) Get(id string) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var t models.Task
	err := m.col.FindOne(ctx, bson.M{"id": id}).Decode(&t)
	if err != nil {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

// Update apenas para as chaves recebidas
func (m *MongoStore) Update(id string, patch map[string]interface{}) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	update := bson.M{}
	if v, ok := patch["title"]; ok {
		if s, ok := v.(string); ok {
			update["title"] = s
		}
	}
	if v, ok := patch["description"]; ok {
		if s, ok := v.(string); ok {
			update["description"] = s
		}
	}
	if v, ok := patch["status"]; ok {
		if s, ok := v.(string); ok {
			update["status"] = s
		}
	}
	if v, ok := patch["priority"]; ok {
		if s, ok := v.(string); ok {
			update["priority"] = s
		}
	}
	if v, ok := patch["due_date"]; ok {
		switch vv := v.(type) {
		case time.Time:
			update["due_date"] = vv
		case *time.Time:
			if vv != nil {
				update["due_date"] = *vv
			}
		case string:
			if parsed, err := models.ParseDate(vv); err == nil {
				update["due_date"] = parsed
			}
		}
	}

	now := time.Now().UTC()
	update["updated_at"] = now

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.Task

	err := m.col.FindOneAndUpdate(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": update},
		opts,
	).Decode(&updated)

	if err != nil {
		return models.Task{}, ErrNotFound
	}

	return updated, nil
}

func (m *MongoStore) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	result, err := m.col.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the MongoDB connection.
func (m *MongoStore) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
