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

func (m SessionModel) Insert(session *Session) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channelObjectID, err := primitive.ObjectIDFromHex(session.ChannelId)
	if err != nil {
		return "", err
	}

	sessionDoc := bson.M{
		"channel_id": channelObjectID,
		"messages":   session.Messages,
		"context":    session.Context,
		"is_root":    session.IsRoot,
	}

	if !session.IsRoot {
		parentObjectID, err := primitive.ObjectIDFromHex(session.ChannelId)
		if err != nil {
			return "", err
		}
		sessionDoc["parent_id"] = parentObjectID
	}

	res, err := m.Collection.InsertOne(ctx, sessionDoc)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return "", ErrCannotInsert
		default:
			return "", err
		}
	}
	session.ID = res.InsertedID.(primitive.ObjectID)
	return session.ID.Hex(), nil
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

func (m SessionModel) Update(id string, session *Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	sessionDoc := bson.M{
		"$push": bson.M{
			"messages": bson.M{
				"$each": session.Messages,
			},
		},
	}

	_, err = m.Collection.UpdateOne(ctx, bson.M{"_id": objectId}, sessionDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (m SessionModel) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = m.Collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}
