package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"misc.sahilsasane.net/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	Name         string               `bson:"name" json:"name"`
	Email        string               `bson:"email" json:"email"`
	PasswordHash []byte               `bson:"password_hash" json:"-"`
	Password     password             `bson:"-" json:"-"`
	Channels     []primitive.ObjectID `bson:"channels" json:"channels"`
	Activated    bool                 `bson:"activated" json:"activated"`
	Version      int                  `bson:"version" json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	Collection *mongo.Collection
}

func (m UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user.CreatedAt = time.Now()
	user.Version = 1

	userDoc := bson.M{
		"created_at":    user.CreatedAt,
		"name":          user.Name,
		"email":         user.Email,
		"password_hash": user.Password.hash,
		"channels":      []primitive.ObjectID{},
		"activated":     user.Activated,
		"version":       user.Version,
	}

	res, err := m.Collection.InsertOne(ctx, userDoc)
	if err != nil {
		switch {
		case mongo.IsDuplicateKeyError(err):
			return ErrDuplicateEmail
		default:
			return nil
		}
	}

	user.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	user.Password.hash = user.PasswordHash

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := m.Collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"name":      user.Name,
				"email":     user.Email,
				"activated": user.Activated,
				"version":   user.Version + 1,
			},
			"$push": bson.M{
				"channels": user.Channels,
			},
		},
	)

	if result.Err() != nil {
		switch {
		case mongo.IsDuplicateKeyError(result.Err()):
			return ErrDuplicateEmail
		case errors.Is(result.Err(), mongo.ErrNoDocuments):
			return ErrEditConflict
		default:
			return result.Err()
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// First find the token
	var token Token
	err := m.Collection.Database().Collection("tokens").FindOne(ctx, bson.M{
		"hash":   tokenHash[:],
		"scope":  tokenScope,
		"expiry": bson.M{"$gt": time.Now()},
	}).Decode(&token)

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Then get the associated user
	var user User
	err = m.Collection.FindOne(ctx, bson.M{"_id": token.UserID}).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Get(id primitive.ObjectID) (*User, error) {
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Collection.FindOne(ctx, bson.M{"_id": primitive.ObjectID(id)}).Decode(&user)

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.Collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return err
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}
