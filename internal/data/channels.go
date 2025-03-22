package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Channel struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id"`
	UserId    string               `json:"user_id" bson:"user_id"`
	Sessions  []primitive.ObjectID `json:"sessions" bson:"sessions"`
	Tree      primitive.ObjectID   `json:"tree" bson:"tree"`
	CreatedAt time.Time            `json:"created_at" bson:"created_at"`
}

type ChannelModel struct {
	Collection *mongo.Collection
}

func (m ChannelModel) Insert(channel *Channel) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channel.CreatedAt = time.Now()

	channelDoc := bson.M{
		"_id":        channel.ID,
		"created_at": time.Now(),
		"tree":       channel.Tree,
		"sessions":   channel.Sessions,
		"user_id":    channel.UserId,
	}

	res, err := m.Collection.InsertOne(ctx, channelDoc)

	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return "", ErrCannotInsert
		default:
			return "", err
		}
	}

	channel.ID = res.InsertedID.(primitive.ObjectID)
	return channel.ID.Hex(), nil
}

func (m ChannelModel) GetById(id string) (*Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var channel Channel

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = m.Collection.FindOne(ctx, bson.M{"_id": primitive.ObjectID(objectID)}).Decode(&channel)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &channel, nil
}

func (m ChannelModel) Update(id string, channel *Channel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	channelDoc := bson.M{
		"$push": bson.M{
			"sessions": bson.M{
				"$each": channel.Sessions,
			},
		},
	}

	err = m.Collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, channelDoc).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}
