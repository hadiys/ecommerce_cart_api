package controllers

import (
	"context"
	"errors"
	"go-com/database"
	"go-com/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prod_collection *mongo.Collection
	user_collection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prod_collection: prodCollection,
		user_collection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc{
	return func(c *gin.Context) {
		// Fetches the query value from the URL parameter keyed "id"
		productQueryID := c.Query("id")
		
		if productQueryID == "" { 
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("Product ID is empty"))
			return 
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("User ID is empty"))
			return 
		}

		// Why do we only generate the objectID for the product but not the user as well?
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err!=nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProductToCart(ctx, app.prod_collection, app.user_collection, productID, userQueryID)
		if err!=nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successfully added to the cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("Product ID is empty"))
			return 
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("User ID is empty"))
			return 
		}

		// Why do we only generate the objectID for the product but not the user as well?
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err!=nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(ctx, app.prod_collection, app.user_collection, productID, userQueryID)
		if err!=nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successfully removed from the cart")
	}
}	

func GetItemFromCart() gin.HandlerFunc{
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
			c.Abort()
			return 
		}

		user_id2, _ := primitive.ObjectIDFromHex(user_id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledCart models.User

		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: user_id2}}).Decode(&filledCart)
		if err!=nil {
			log.Println(err)
			c.IndentedJSON(500, "ID not found")
			return 
		}

		filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_id2}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		if err!=nil {
			log.Println(err)
		}

		var listing []bson.M
		if err = pointCursor.All(ctx, &listing); err!=nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledCart.UserCart)
		}
		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc{
		return func(c *gin.Context){
			userQueryID := c.Query("id")
			if userQueryID == "" {
				log.Panicln("User ID is empty")
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("User ID is empty"))
				return 
			}
			var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := database.BuyItemFromCart(ctx, app.user_collection, userQueryID)
			if err!=nil {
				c.IndentedJSON(http.StatusInternalServerError, err)
			}
			c.IndentedJSON(200, "Successfully placed the order")
		}
}

func (app *Application) InstantBuy() gin.HandlerFunc{
	return func(c *gin.Context){
		userQueryID := c.Query("id")
		if userQueryID == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("User ID is empty"))
			return 
		}

		productQueryID := c.Query("pid")
		if productQueryID == "" {
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("Product ID is empty"))
			return 
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err!=nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuyer(ctx, app.prod_collection, app.user_collection, productID, userQueryID)
		if err!=nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successfully placed the order")
	}
}
