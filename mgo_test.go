package mgocli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _mgocli_ *MgoCli
var _testDatabase_ = "mgocli_test"
var _testCollection_ = "users"
var _cluster_ = "mongodb://mongo1:27017,mongo2:27018,mongo3:27019/?replicaSet=myReplicaSet"

// https://www.mongodb.com/resources/products/compatibilities/deploying-a-mongodb-cluster-with-docker

type user struct {
	Id   *bson.ObjectID `bson:"_id,omitempty"`
	Uid  int            `bson:"uid"`
	Name string         `bson:"name"`
	Age  int            `bson:"age"`
	Data string         `bson:"data,omitempty"`
}

func TestMain(m *testing.M) {
	_mgocli_, _ = New(_cluster_, _testDatabase_)
	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	if err := _mgocli_.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCollection(t *testing.T) {
	if err := _mgocli_.CreateCollection(context.Background(), _testCollection_); err != nil {
		t.Fatal(err)
	}
}

func TestCreateIndex(t *testing.T) {
	index := mongo.IndexModel{
		Keys:    bson.D{{"uid", -1}},
		Options: options.Index().SetUnique(true),
	}
	if err := _mgocli_.CreateIndex(context.Background(), _testCollection_, index); err != nil {
		t.Fatal(err)
	}
}

func TestInsert(t *testing.T) {
	ret, err := _mgocli_.InsertOne(context.TODO(), _testCollection_, &user{
		Uid:  2,
		Name: "Bob",
		Age:  18,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ret)
}

func TestInsertMany(t *testing.T) {
	us := []user{{
		Uid:  3,
		Name: "Alice",
		Age:  18,
	}, {
		Uid:  4,
		Name: "Bob",
		Age:  18,
	}}
	ret, err := _mgocli_.InsertMany(context.TODO(), _testCollection_, us)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ret)
}

func TestUpdateById(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("67060f70e0f81dedad0333dd")

	update := bson.D{{"$set", bson.D{{"age", 23}}}}

	if err := _mgocli_.UpdateById(context.Background(), _testCollection_, id, update); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	filter := bson.D{{"uid", 1}}

	update := bson.D{{"$set", bson.D{{"age", 19}}}}

	if _, err := _mgocli_.Updates(context.Background(), _testCollection_, filter, update); err != nil {
		t.Fatal(err)
	}
}

func TestFindOne(t *testing.T) {
	filter := bson.D{{"uid", 1}}
	var u user
	b, err := _mgocli_.FindOne(context.Background(), _testCollection_, filter, &u)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
	t.Log(u)
}

func TestFind(t *testing.T) {
	filter := bson.D{}
	var users []user
	if err := _mgocli_.Find(context.Background(), _testCollection_, filter, &users); err != nil {
		t.Fatal(err)
	}
	t.Log(users)
}

func TestDeleteById(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("67060f70e0f81dedad0333dd")
	_mgocli_.DeleteById(context.Background(), _testCollection_, id)
}

func TestTransaction(t *testing.T) {
	// s
	err := _mgocli_.Transaction(context.Background(), func(ctx context.Context) (interface{}, error) {
		id, _ := bson.ObjectIDFromHex("6706111ca325ca042c711e7d")
		if err := _mgocli_.UpdateById(ctx, _testCollection_, id, bson.D{{"$set", bson.D{{"uid", 5}}}}); err != nil {
			return nil, err
		}
		rlt, err := _mgocli_.Updates(ctx, _testCollection_, bson.D{{"uid", 5}}, bson.D{{"$set", bson.D{{"age", 100}}}})
		if err != nil {
			return nil, err
		}
		if rlt.ModifiedCount != 1 {
			return nil, errors.New("rlt.ModifiedCount is not 1")
		}

		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
