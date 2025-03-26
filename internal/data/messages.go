package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	SessionId string             `json:"session_id" bson:"session_id"`
	Data      struct {
		Role  string        `json:"role"`
		Parts []interface{} `json:"parts"`
	} `json:"data"`
}

type MessageModel struct {
	Collection *mongo.Collection
}

func (m MessageModel) Insert(message *Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sessionObjectId, err := primitive.ObjectIDFromHex(message.SessionId)
	if err != nil {
		return err
	}

	messageDoc := bson.M{
		"session_id": sessionObjectId,
		"data":       message.Data,
	}

	_, err = m.Collection.InsertOne(ctx, messageDoc)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return ErrCannotInsert
		default:
			return err
		}
	}

	return nil
}

func (m MessageModel) GetById(id string) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var message Message
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = m.Collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&message)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &message, nil
}
