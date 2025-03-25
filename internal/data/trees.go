package data

import (
	"context"
	"fmt"
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

	objectID, err := primitive.ObjectIDFromHex(tree.ChannelId)
	if err != nil {
		return "", err
	}

	treeDoc := bson.M{
		"channel_id": objectID,
		"root":       tree.Root,
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
	fmt.Print(objectID)
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
			"root":       tree.Root,
			"updated_at": time.Now(),
			"tree":       tree.TreeStructure,
		},
	}
	fmt.Print("\ndone\n\n")

	err = m.Collection.FindOneAndUpdate(ctx, bson.M{"channel_id": objectID}, treeDoc).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRecordNotFound
		}
		return err
	}
	fmt.Print("\ndone\n\n")
	return nil
}
