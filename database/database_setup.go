package database

import (
	"context"
	"log"
	// "reflect"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetTimeout(10*time.Second)
	// fmt.Println("ClientOption type: ", reflect.TypeOf(clientOptions))

	// client, err := mongo.Connect(context.TODO(), clientOptions)
	// if err!=nil {
	// 	panic(err)
	// }

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err!=nil {
	// 		panic(err)
	// 	}
	// }()

	// err = client.Ping(context.TODO(), nil)
	// if err!=nil {
	// 	log.Println("Failed to connect to MongoDB")
	// 	return nil
	// }

	// fmt.Println("Connected to MongoDB")
	// return client


	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("failed to connect to mongodb")
		return nil
	}
	fmt.Println("Successfully Connected to the mongodb")
	return client

}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return collection 
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return collection
}