package controllers

import (
	"context"
	"fmt"
	"go-com/database"
	"go-com/models"
	"go-com/tokens"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProdCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	
	return string(bytes)
}

func VerifyPassword(userPassword, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true 
	msg := ""

	if err!=nil {
		msg = "Email or password are incorrect"
		valid = false
	}

	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already in use"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := tokens.GenerateToken(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.Refresh_Token = &refreshToken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)

		_, insertErr := UserCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not inserted"})
			return
		}

		defer cancel()
		c.JSON(http.StatusCreated, "Successfully signed up!")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user, foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			// Is this the correct error message?
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login or Password incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
		}

		token, refreshToken,  _ := tokens.GenerateToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		defer cancel()
		
		// What is the difference between GenerateToken & UpdateAllTokens?
		tokens.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		c.JSON(http.StatusFound, foundUser)
	}

}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var product models.Product
		defer cancel()

		// Not sure what is the use of BindJSON here
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Creating a new ID for the product and inserting it into the DB
		product.Product_id = primitive.NewObjectID()
		_, anyerr := ProdCollection.InsertOne(ctx, product)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not created"})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, "Successfully added Product Admin.")

	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Retrieve all products
		cursor, err := ProdCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)
		// Returns the last error seen by the cursor, otherwise it returns nil
		if err := cursor.Err(); err != nil {
			// Don't forget to log errors. I log them really simple here just
			// to get the point across.
			log.Println(err)
			c.IndentedJSON(400, "Invalid")
			return
		}

		defer cancel()
		c.IndentedJSON(200, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("Query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid index search"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDB, err := ProdCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
		if err != nil {
			c.IndentedJSON(400, "Something went wrong while fetching the DB query")
			return
		}

		err = searchQueryDB.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid")
			return
		}

		// Closes the cursor
		defer searchQueryDB.Close(ctx)

		if err := searchQueryDB.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid request")
			return
		}

		defer cancel()
		c.IndentedJSON(200, searchProducts)

	}
}
