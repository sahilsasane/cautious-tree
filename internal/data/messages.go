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
		Role  string              `json:"role"`
		Parts []map[string]string `json:"parts"`
	} `json:"data"`
}

type MessageModel struct {
	Collection *mongo.Collection
}

func (m MessageModel) Insert(message *Message) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sessionObjectId, err := primitive.ObjectIDFromHex(message.SessionId)
	if err != nil {
		return "", err
	}

	messageDoc := bson.M{
		"session_id": sessionObjectId,
		"data":       message.Data,
	}

	res, err := m.Collection.InsertOne(ctx, messageDoc)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return "", ErrCannotInsert
		default:
			return "", err
		}
	}
	message.ID = res.InsertedID.(primitive.ObjectID)

	return message.ID.Hex(), nil
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

func (m MessageModel) GetAllMesssageById(ids []primitive.ObjectID) ([]*Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := m.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, nil
	}

	return messages, nil
}
