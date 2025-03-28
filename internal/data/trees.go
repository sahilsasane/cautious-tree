package data

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Tree struct {
	ID            primitive.ObjectID     `json:"id" bson:"_id"`
	ChannelId     string                 `json:"channel_id" bson:"channel_id"`
	Root          string                 `json:"root" bson:"root"`
	CreatedAt     time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" bson:"updated_at"`
	TreeStructure map[string]interface{} `json:"tree" bson:"tree"`
}

type TreeModel struct {
	Collection *mongo.Collection
}

func (m TreeModel) Insert(tree *Tree) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channelObjectID, err := primitive.ObjectIDFromHex(tree.ChannelId)
	if err != nil {
		return "", err
	}

	var rootValue interface{} = nil

	if tree.Root != "" {
		rootValue, err = primitive.ObjectIDFromHex(tree.Root)
		if err != nil {
			return "", err
		}
	}

	treeDoc := bson.M{
		"channel_id": channelObjectID,
		"root":       rootValue,
		"tree":       tree.TreeStructure,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	res, err := m.Collection.InsertOne(ctx, treeDoc)

	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return "", ErrDuplicateEmail
		default:
			return "", err
		}
	}
	tree.ID = res.InsertedID.(primitive.ObjectID)

	return tree.ID.Hex(), nil
}

func (m TreeModel) GetByChannelId(id string) (*Tree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tree Tree

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = m.Collection.FindOne(ctx, bson.M{"channel_id": objectID}).Decode(&tree)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &tree, nil
}

func (m TreeModel) Update(id string, tree *Tree) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	treeDoc := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"tree":       tree.TreeStructure,
		},
	}

	if tree.Root != "" {
		rootObjectId, err := primitive.ObjectIDFromHex(tree.Root)
		if err != nil {
			return err
		}
		treeDoc["$set"].(bson.M)["root"] = rootObjectId
	}

	err = m.Collection.FindOneAndUpdate(ctx, bson.M{"channel_id": objectID}, treeDoc).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}
