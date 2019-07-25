package users

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       *primitive.ObjectID `json:"id" bson:"_id"`
	Name     string              `json:"name,omitempty" bson:"name,omitempty"`
	Email    string              `json:"email" bson:"email"`
	Password string              `json:"password,omitempty" bson:"password"`
}

func GetUser(ctx context.Context, client *mongo.Client, id string) (*User, error) {
	coll := client.Database("codev").Collection("users")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrUserNotExists
	}
	singleRes := coll.FindOne(ctx, bson.D{{"_id", objId}})
	if singleRes.Err() != nil {
		if singleRes.Err() == mongo.ErrNoDocuments {
			return nil, ErrUserNotExists
		}
		return nil, singleRes.Err()
	}
	var user User
	err = singleRes.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotExists
		}
		return nil, err
	}
	user.Password = ""
	return &user, nil
}