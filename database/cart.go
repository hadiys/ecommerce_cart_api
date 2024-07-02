package database

import(
	"context"
	"errors"
	"log"
	"time"

	"go-com/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

)

var (
	ErrCantFindProduct = errors.New("Can't find product")
	ErrCantDecodeProducts = errors.New("Can't decode product")
	ErrUserIDIsNotValid = errors.New("User ID is not valid")
	ErrCantUpdateUser = errors.New("Cannot add item to cart")
	ErrCantRemoveItem = errors.New("Cannot remove item from cart")
	ErrCantGetItem = errors.New("Cannot get item from cart")
	ErrCantBuyCartItem = errors.New("Cannot update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	
	//Returns a Cursor for matching documents. Looking for thr product by its ID
	searchFromDB, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err!=nil{
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.ProductUser 	
	
	//Iterates the cursor and decodes each doc into a result.
	err = searchFromDB.All(ctx, &productCart)
	if err!=nil{
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err!=nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	
	// When the product is found in the db, decoded successfully into golang, 
	// and userID is validated, an item can be added to the user's cart
	// DB needs to be updated, giving it the id of the user and the product written to productCart

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err!=nil {
		log.Println(err)
		return ErrCantUpdateUser
	}

	return nil
}

func RemoveCartItem(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	// Validate the userID
	id, err := primitive.ObjectIDFromHex(userID)
	if err!=nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}

	// Removing item from User's cart using the productID
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err!=nil {
		log.Println(err)
		return ErrCantRemoveItem
	}
	
	return nil
}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error{
	
	// The caller of this function passes userID as a hex string which needs to be validated
	id, err := primitive.ObjectIDFromHex(userID)
	if err!=nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	
	// A User
	var getCartItems models.User

	// An Order
	var orderCart models.Order

	// For that Order we are creating an ID, time ordered, a slice of ProductUser, and payment method
	// []ProductUser is the same as []Products.. Why not just use []Product? 
	orderCart.Order_id = primitive.NewObjectID()
	orderCart.Ordered_at = time.Now()
	orderCart.Order_cart = make([]models.ProductUser, 0)
	orderCart.Payment_method.COD = true 


	// $usercart is the field representing the slice of Products of a User. 
	// Seems like they are being retrieved from the DB as individual items using $unwind
	unwind 	:= bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}} }}
	grouping:= bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

	// Aggregation result from DB, probably returning the order price as prices of individual products
	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})

	// Obtains a channel that gets closed when the ctx is cancelled or timed out
	ctx.Done()

	if err!=nil {
		panic(err)
	}

	// Decoding the result into getUserCart by iterating over currentResults
	// bson.M is handled in Go as: map[string]interface{} 
	// Hence can be iterated over using for-range loop
	var getUserCart []bson.M 
	if err = currentResults.All(ctx, &getUserCart); err!=nil {
		panic(err)
	}

	// Save the total price of items in the cart 
	var total_price int32 
	for _, user_item := range getUserCart {
		price := user_item["total"]
		total_price = price.(int32) //Shouldnt it be total_price += price.(int32)
									// Actually grouping stage in the DB already totaled the price of the items
	}								// So why do we iterate over a result that already has the total?

	// Update the order's total price
	orderCart.Price = int(total_price)

	// Update the User's Order in the DB
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err!=nil {
		log.Println(err)
	}

	// Retrieving the items added to the cart from the DB and decoding it into the user's cart
	err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err!=nil {
		log.Println(err) 
	}

	// Updating the User's []Order with the items in the User's cart
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err!=nil {
		log.Println(err)
	}

	// Empty the user's cart to complete the purchase
	usercart_empty := make([]models.ProductUser, 0)
	filtered := bson.D{primitive.E{Key: "_id", Value: id}}
	updated := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key:"usercart", Value: usercart_empty}}}}
	_, err = userCollection.UpdateOne(ctx, filtered, updated)
	if err!=nil {
		return ErrCantBuyCartItem
	}

	return nil
}

func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err!=nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}

	// ProductUser is identical to Product aside from datatypes. 
	// Not sure why they should both exist
	var product_details models.ProductUser
	var order_details models.Order 

	order_details.Order_id = primitive.NewObjectID()
	order_details.Ordered_at = time.Now()
	order_details.Order_cart = make([]models.ProductUser, 0)
	order_details.Payment_method.COD = true 

	// Retrieving the product from the DB and saving it to product_details
	err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&product_details)
	if err!=nil {
		log.Println(err)
	}

	order_details.Price = product_details.Price

	// Updating the User's Orders with order_details
	filter := bson.D{primitive.E{Key:"_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: order_details}}}}
	_, err = userCollection.UpdateOne(ctx, filter ,update)
	if err!=nil {
		log.Println(err)
	}

	// Updating the Order's cart with product_details
	filter2 := bson.D{primitive.E{Key:"_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": product_details}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err!=nil {
		log.Println(err)
	}

	return nil 
}