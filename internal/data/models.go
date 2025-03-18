package data

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users UserModel
}

func NewModels(client *mongo.Client, dbName string) Models {
	db := client.Database(dbName)

	return Models{
		Users: UserModel{Collection: db.Collection("users")},
	}
}
