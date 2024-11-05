package mgocli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type MgoCli struct {
	client   *mongo.Client
	database string
}

func New(dial string, database string) (*MgoCli, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(dial))
	return &MgoCli{client: client, database: database}, err
}

func (mc *MgoCli) Close(ctx context.Context) {
	_ = mc.client.Disconnect(ctx)
}

func (mc *MgoCli) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return mc.client.Ping(ctx, readpref.Primary())
}

// CreateCollection if database not exist,it will create it
func (mc *MgoCli) CreateCollection(ctx context.Context, collection string) error {
	return mc.client.Database(mc.database).CreateCollection(ctx, collection)
}

func (mc *MgoCli) CreateIndex(ctx context.Context, collection string, index mongo.IndexModel) error {
	_, err := mc.client.Database(mc.database).Collection(collection).Indexes().CreateOne(ctx, index)
	return err
}

func (mc *MgoCli) FindOne(ctx context.Context, collection string, filter interface{}, document interface{}) (bool, error) {
	err := mc.client.Database(mc.database).Collection(collection).FindOne(ctx, filter).Decode(document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (mc *MgoCli) Find(ctx context.Context, collection string, filter interface{}, documents interface{}) error {
	cursor, err := mc.client.Database(mc.database).
		Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, documents); err != nil {
		return err
	}
	return nil
}

func (mc *MgoCli) FindPages(ctx context.Context, collection string, filter interface{}, documents interface{}, sort interface{}, skip, limit int64) error {
	cursor, err := mc.client.Database(mc.database).
		Collection(collection).Find(ctx, filter,
		options.Find().SetSort(sort).SetSkip(skip).SetLimit(limit))
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, documents); err != nil {
		return err
	}
	return nil
}

// InsertOne
func (mc *MgoCli) InsertOne(ctx context.Context, collection string, document interface{}) (bool, error) {
	rlt, err := mc.client.Database(mc.database).
		Collection(collection).
		InsertOne(ctx, document)
	if err != nil || !rlt.Acknowledged {
		return false, err
	}
	return true, err
}

func (mc *MgoCli) MustInsertOne(ctx context.Context, collection string, document interface{}) error {
	rlt, err := mc.client.Database(mc.database).
		Collection(collection).
		InsertOne(ctx, document)
	if err != nil {
		return err
	}
	if !rlt.Acknowledged {
		return fmt.Errorf("could not insert document")
	}
	return nil
}

func (mc *MgoCli) InsertMany(ctx context.Context, collection string, documents interface{}) (int, error) {
	rlt, err := mc.client.Database(mc.database).
		Collection(collection).
		InsertMany(ctx, documents)
	if err != nil {
		return 0, err
	}
	return len(rlt.InsertedIDs), err
}

// MustUpdateById
func (mc *MgoCli) MustUpdateById(ctx context.Context, collection string, id bson.ObjectID, update bson.D) error {
	rlt, err := mc.updateById(ctx, collection, id, update)
	if err != nil {
		return err
	}
	if rlt.ModifiedCount != 1 {
		return fmt.Errorf("wrong ModifiedCount %d", rlt.ModifiedCount)
	}
	return nil
}

// UpdateById
func (mc *MgoCli) UpdateById(ctx context.Context, collection string, id bson.ObjectID, update bson.D) error {
	_, err := mc.updateById(ctx, collection, id, update)
	return err
}

func (mc *MgoCli) updateById(ctx context.Context, collection string, id bson.ObjectID, update bson.D) (*mongo.UpdateResult, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	return mc.client.Database(mc.database).
		Collection(collection).
		UpdateOne(ctx, filter, update)
}

func (mc *MgoCli) Updates(ctx context.Context, collection string, filter bson.D, update bson.D) (*mongo.UpdateResult, error) {
	rlt, err := mc.client.Database(mc.database).
		Collection(collection).
		UpdateMany(ctx, filter, update)

	return rlt, err
}

func (mc *MgoCli) MustDeleteById(ctx context.Context, collection string, id bson.ObjectID) error {
	rlt, err := mc.deleteById(ctx, collection, id)
	if err != nil {
		return err
	}
	if rlt.DeletedCount != 1 {
		return fmt.Errorf("wrong DeletedCount %d", rlt.DeletedCount)
	}
	return err
}

func (mc *MgoCli) DeleteById(ctx context.Context, collection string, id bson.ObjectID) error {
	_, err := mc.deleteById(ctx, collection, id)
	return err
}

func (mc *MgoCli) deleteById(ctx context.Context, collection string, id bson.ObjectID) (*mongo.DeleteResult, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	return mc.client.Database(mc.database).
		Collection(collection).
		DeleteOne(ctx, filter)
}

func (mc *MgoCli) Deletes(ctx context.Context, collection string, filter bson.D) (*mongo.DeleteResult, error) {
	return mc.client.Database(mc.database).
		Collection(collection).
		DeleteMany(ctx, filter)
}

func (mc *MgoCli) Transaction(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) error {
	// https://www.mongodb.com/zh-cn/docs/manual/core/transactions/#read-concern-write-concern-read-preference
	txnOptions := options.Transaction().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.Primary())

	session, err := mc.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, fn, txnOptions)

	return err
}
