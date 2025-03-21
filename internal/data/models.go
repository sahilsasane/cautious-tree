package data

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrCannotInsert   = errors.New("cannot insert record")
)

type Models struct {
	Users    UserModel
	Tokens   TokenModel
	Sessions SessionModel
	Messages MessageModel
	Trees    TreeModel
	Channel  ChannelModel
}

func NewModels(client *mongo.Client, dbName string) Models {
	db := client.Database(dbName)
	users := UserModel{Collection: db.Collection("users")}

	// Create indexes on startup
	if err := users.CreateIndexes(); err != nil {
		panic(err) // In a production app, you might want to handle this error differently
	}

	return Models{
		Users:   users,
		Tokens:  TokenModel{Collection: db.Collection("tokens")},
		Channel: ChannelModel{Collection: db.Collection("channels")},
		Trees:   TreeModel{Collection: db.Collection("trees")},
	}
}
