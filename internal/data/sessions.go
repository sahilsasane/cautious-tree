package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Session struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id"`
	ChannelId string               `json:"channel_id" bson:"channel_id"`
	Messages  []primitive.ObjectID `json:"messages" bson:"messages"`
	Context   string               `json:"context" bson:"context"`
	IsRoot    bool                 `json:"is_root" bson:"is_root"`
	ParentId  string               `json:"parent_id" bson:"parent_id"`
}

type SessionModel struct {
	Collection *mongo.Collection
}

func (m SessionModel) Insert(session *Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sessionDoc := bson.M{
		"channel_id": session.ChannelId,
		"messages":   session.Messages,
		"context":    session.Context,
		"is_root":    session.IsRoot,
		"parent_id":  session.ParentId,
	}

	_, err := m.Collection.InsertOne(ctx, sessionDoc)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return ErrDuplicateEmail
		default:
			return nil
		}
	}

	return nil
}

func (m SessionModel) GetById(id string) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var session Session
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = m.Collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&session)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &session, nil
}
