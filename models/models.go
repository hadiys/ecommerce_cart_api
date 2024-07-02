package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID				primitive.ObjectID			`json:"_id" bson:"_id"`
	First_name		*string						`json:"first_name" validate:"required,min=2,max=30"`
	Last_name		*string						`json:"last_name" validate:"required,min=2,max=30"`
	Password		*string						`json:"password" validate:"required,min=6"`
	Email			*string						`json:"email" validate:"email,required"`
	Phone			*string						`json:"phone" validate:"required"`
	Token			*string						`json:"token"`
	Refresh_Token	*string						`json:"refresh_token"`
	Created_at		time.Time 					`json:"created_at"`
	Updated_at		time.Time 					`json:"updated_at"`
	User_id			string 						`json:"user_id"`
	UserCart		[]ProductUser 				`json:"usercart" bson:"usercart"`		
	Address_Details	[]Address  					`json:"address" bson:"address"`
	Order_Status	[]Order 					`json:"orders" bson:"orders"`
}

type Product struct {
	Product_id			primitive.ObjectID		 `json:"_id" bson:"_id"`
	Product_name		*string 			   	 `json:"product_name"`		
	Price				*uint64 			   	 `json:"price"`
	Rating				*uint8  			   	 `json:"rating"`
	Image				*string  			   	 `json:"image"`
}

type ProductUser struct {
	Product_id			primitive.ObjectID 		 `json:"_id" bson:"_id"`
	Product_name		*string 			   	 `json:"product_name" bson:"product_name"` 
	Price				int  			   	 	 `json:"price" bson:"price"`
	Rating				*uint64  			   	 `json:"rating" bson:"rating"`
	Image				*string	 			   	 `json:"image" bson:"image"`
}

type Address struct {
	Address_id			primitive.ObjectID 		 `json:"_id" bson:"_id"`
	House				*string 				 `json:"house_name" bson:"house_name"`
	Street				*string  				 `json:"street_name" bson:"street_name"`
	City				*string 				 `json:"city_name" bson:"city_name"`
	PostCode			*string 				 `json:"postcode" bson:"postcode"`
}

type Order struct {
	Order_id			primitive.ObjectID 		 `json:"_id" bson:"_id"` 
	Order_cart			[]ProductUser 		 	 `json:"order_list" bson:"order_list"`
	Ordered_at			time.Time  		 		 `json:"ordered_at" bson:"ordered_at"`
	Price				int 		 			 `json:"total_price" bson:"total_price"`
	Discount			*int 		  			 `json:"discount" bson:"discount"` 
	Payment_method		Payment  		 		 `json:"payment_method" bson:"payment_method"`
}

type Payment struct {
	Digital bool
	COD 	bool
}