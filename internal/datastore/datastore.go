package datastore

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"gitlab.com/PROD/go-service-boilerplate/internal/v1/models"
)

type Repository interface {
	Close()
	EnsureIndexes(coll string, indexQuery []string)
	Find(query Query) (*[]models.Model, error)
	Insert(query Query, d interface{}) (interface{}, error)
	Update(query Query, d interface{}) (interface{}, error)
	Delete(query Query) (interface{}, error)
	Paginate(query Query, page Pagination) (*[]models.Model, error)
	Aggregate(query Query, pipeline []bson.M) (*mongo.Cursor, error)
}

// Config Is the Datastore config
type Config struct {
	Uri          string
	DatabaseName string
}

// Query Is the query builder object
type Query struct {
	Select bson.M
	Where  bson.M
	From   string
}

type Pagination struct {
	Page  int
	Limit int
	Sort  bson.M
}

type datastore struct {
	c  *mongo.Client
	db *mongo.Database
}

// NewDatastore Will initialize a new datastore which contains the client connection
func NewDatastore(config *Config) Repository {
	client, err := mongo.NewClient(options.Client().ApplyURI(config.Uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(config.DatabaseName)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return &datastore{c: client, db: db}
}

// Close Will close the datastores connection
func (ds *datastore) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := ds.c.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// EnsureIndexes Makes sure datastore indexes are created
func (ds *datastore) EnsureIndexes(coll string, indexQuery []string) {
	opts := options.CreateIndexes().SetMaxTime(5 * time.Second)
	index := []mongo.IndexModel{}
	for _, val := range indexQuery {
		tmp := mongo.IndexModel{
			Options: options.Index().SetUnique(true),
		}
		tmp.Keys = bsonx.Doc{{Key: val, Value: bsonx.Int32(1)}}
		index = append(index, tmp)
	}

	_, err := ds.db.Collection(coll).Indexes().CreateMany(context.Background(), index, opts)
	if err != nil {
		log.Fatal(err)
	}
}

// Find Will find an entry within the datastore
func (ds *datastore) Find(query Query) (*[]models.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	o := options.Find().SetProjection(query.Select)
	cursor, err := ds.db.Collection(query.From).Find(ctx, query.Where, o)
	if err != nil {
		return nil, err
	}

	var res []models.Model
	for cursor.Next(context.Background()) {
		var m models.Model
		_ = cursor.Decode(&m)
		res = append(res, m)
	}

	return &res, err
}

// Insert Will insert an entry into datastore
func (ds *datastore) Insert(query Query, d interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := ds.db.Collection(query.From).InsertOne(ctx, d)
	return res, err
}

// Update will update an entry from the datastore
func (ds *datastore) Update(query Query, d interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res := ds.db.Collection(query.From).FindOneAndUpdate(ctx, query.Where, d)
	return res, res.Err()
}

// Delete will delete an entry from the datastore
func (ds *datastore) Delete(query Query) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res := ds.db.Collection(query.From).FindOneAndDelete(ctx, query.Where)
	return res, res.Err()
}

// Paginate provides pagination to the find operation
func (ds *datastore) Paginate(query Query, page Pagination) (*[]models.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	o := options.Find().SetSort(page.Sort).SetLimit(int64(page.Limit)).SetSkip(int64((page.Page - 1) * page.Limit))
	cursor, err := ds.db.Collection(query.From).Find(ctx, query.Where, o)
	if err != nil {
		return nil, err
	}

	var res []models.Model
	for cursor.Next(context.Background()) {
		var m models.Model
		_ = cursor.Decode(&m)
	}

	return &res, err
}

// Aggregate uses mongodbs Aggregate operation
func (ds *datastore) Aggregate(query Query, pipeline []bson.M) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return ds.db.Collection(query.From).Aggregate(ctx, pipeline)
}
