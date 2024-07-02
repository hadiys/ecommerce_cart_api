package tokens

import (
	"context"
	"log"
	"os"
	"time"

	"go-com/database"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)	

type SignedDetails struct {
	Email		string
	First_name 	string
	Last_name	string
	Uid			string
	jwt.StandardClaims
}

var UserData *mongo.Collection = database.UserData(database.Client, "Users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateToken(email, first_name, last_name, uid string) (signedToken, signedRefreshToken string, err error) {
	
	// Generating a token as an instance of SignedDetails with an expiry of 24hrs
	claims := &SignedDetails{
		Email: email,
		First_name: first_name,
		Last_name: last_name,
		Uid: uid, 
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	// Generating a Refresh token, only defines expiry. Not sure why.
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	// Not sure what is happening from here on
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err!=nil {
		return "", "", nil
	}

	refreshtoken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err!=nil {
		log.Panicln(err)
		return 
	}

	return token, refreshtoken, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token)(interface{}, error) {
		return []byte(SECRET_KEY), nil 
	})

	if err!=nil {
		msg = err.Error()
		return 
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "Token is invalid"
		return 
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "Token is expired"
		return 
	}

	return claims, msg 
}

func UpdateAllTokens(signedToken, signedRefreshToken, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var updateObj primitive.D 

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedToken})
	
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

	// Upsert: when true a document will be inserted if no documents match the filter
	// Seems to be an operation related to new users
	// But it doesnt make sense because why didnt we initialise the created_at field of the user? 
	upsert := true 
	filter := bson.M{"user_id": userID}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	//Update the user's data in the DB
	_, err := UserData.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateObj},
	},
	&opt,
	)
	
	defer cancel()
	if err!=nil {
		log.Panic(err)
		return 
	}
}