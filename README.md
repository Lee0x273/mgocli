
# MongoDB Client for Go - README
## Overview
This project provides a wrapper around `go.mongodb.org/mongo-driver/v2` to facilitate easier usage of MongoDB in Go applications. The goal is to simplify common operations and enhance productivity when working with MongoDB.
## Features
- Simplified connection setup
- Easy CRUD operations
- Helper functions for common tasks
- Error handling and logging
## Requirements
- Go version 1.15 or higher
- MongoDB server running and accessible
## Installation
To install the MongoDB driver wrapper, run the following command:
```bash
go get github.com/Lee0x273/mgocli
```
Replace `yourusername` with your actual GitHub username where the repository is hosted.
## Usage
### Importing the Package
```go
import "github.com/Lee0x273/mgocli"
```
### Setting Up the Connection
```go
package main
import (
	"context"
	"log"
	"github.com/Lee0x273/mgocli"
)
func main() {
	// Set up MongoDB connection
	client, err := mongodbwrapper.NewClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())
	// Use the client for further operations
}
```
### CRUD Operations
#### Create
```go
collection := client.Database("yourdb").Collection("yourcollection")
data := bson.M{"name": "John Doe", "age": 30}
insertResult, err := collection.InsertOne(context.TODO(), data)
if err != nil {
log.Fatal(err)
}
log.Println("Inserted document ID:", insertResult.InsertedID)
```
#### Read
```go
var result bson.M
filter := bson.M{"name": "John Doe"}
err := collection.FindOne(context.TODO(), filter).Decode(&result)
if err != nil {
log.Fatal(err)
}
log.Println("Found document:", result)
```
#### Update
```go
filter := bson.M{"name": "John Doe"}
update := bson.M{"$set": bson.M{"age": 31}}
updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
if err != nil {
log.Fatal(err)
}
log.Println("Modified count:", updateResult.ModifiedCount)
```
#### Delete
```go
filter := bson.M{"name": "John Doe"}
deleteResult, err := collection.DeleteOne(context.TODO(), filter)
if err != nil {
log.Fatal(err)
}
log.Println("Deleted count:", deleteResult.DeletedCount)
```
## Contributing
Contributions to this project are welcome! Please ensure that your code adheres to the Go coding standards and passes all tests before submitting a pull request.
## License
This project is licensed under the MIT License - see the LICENSE file for details.
