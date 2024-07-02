package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	
	"go-com/models"
	
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Returns the keyed url query value
		user_id := c.Query("id")

		// Need to ensure user id is provided in the request before proceeding
		if user_id == "" {
			// Writes a header in the response when providing it with key:value args
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
			c.Abort()
			return
		}

		// Generating an address using the hex string of user_id?
		address, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
		}

		//Creating a new object for the new address
		var addresses models.Address
		addresses.Address_id = primitive.NewObjectID()

		// If the object fails to bind to JSON throw an error
		if err = c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// Filtering documents based on the _id created for address, which is the user_id
		// Flatten the results on the field 'address'
		// Not sure whats happening in the group stage
		match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
		}

		var addressInfo []bson.M

		// Binding the results in addressInfo, throws an error if something was wrong
		if err = pointCursor.All(ctx, &addressInfo); err != nil {
			panic(err)
		}

		// Saving the number into size
		var size int32
		for _, address_no := range addressInfo {
			count := address_no["count"]
			size = count.(int32)
		}

		// Add the address only if there are currently less than 2 addresses.
		// Otherwise respond with an error.
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			c.IndentedJSON(400, "Operation not allowed: Cannot add more than 2 addresses")
		}

		defer cancel()
		ctx.Done()
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		// If user id is not provided in the header, respond with an error.
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "ID not provided"})
			c.Abort()
			return
		}

		user_id2, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, err)
		}

		//If the address cannot bind to JSON respond with an error
		var editAddress models.Address
		if err = c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: user_id2}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editAddress.House}, {Key: "address.0.street_name", Value: editAddress.Street}, {Key: "address.0.city_name", Value: editAddress.City}, {Key: "address.0.postcode", Value: editAddress.PostCode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.IndentedJSON(500, "Something went wrong: Could not update the address")
			return
		}

		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully updated the home address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		// If user id is not provided in the header, respond with an error.
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "ID not provided"})
			c.Abort()
			return
		}

		user_id2, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, err)
		}

		//If the address cannot bind to JSON respond with an error
		var editAddress models.Address
		if err = c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: user_id2}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editAddress.House}, {Key: "address.1.street_name", Value: editAddress.Street}, {Key: "address.1.city_name", Value: editAddress.City}, {Key: "address.1.postcode", Value: editAddress.PostCode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.IndentedJSON(500, "Something went wrong: Could not update the address")
			return
		}

		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully updated the work address")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		user_id2, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		//Deleting all existing addresses?
		filter := bson.D{primitive.E{Key: "_id", Value: user_id2}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.IndentedJSON(404, "Wrong")
			return
		}

		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully deleted (all addresses?)")

	}
}
