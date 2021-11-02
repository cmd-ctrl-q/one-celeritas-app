package data

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

const (
	authorization = "Authorization"
)

var (
	errNoAuthHeader   = errors.New("no authorization header received")
	errTokenWrongSize = errors.New("token is the wrong size")
	errTokenNoMatch   = errors.New("no matching token found")
	errTokenExpired   = errors.New("expired token")
	errUserNoMatch    = errors.New("no matching user found")
)

type Token struct {
	ID        int    `db:"id" json:"id"`
	UserID    int    `db:"user_id" json:"user_id"`
	FirstName string `db:"first_name" json:"first_name"`
	Email     string `db:"email" json:"email"`

	// PlainText stores plain text tokens
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

// GetUserForToken gets a user from the given token
func (t *Token) GetUserForToken(token string) (*User, error) {
	var u User
	var theToken Token

	collection := upper.Collection(t.Table())
	// get token
	res := collection.Find(up.Cond{"token": token})
	err := res.One(&theToken)
	if err != nil {
		return nil, err
	}

	collection = upper.Collection(u.Table())
	// get user from users table
	res = collection.Find(up.Cond{"id": t.UserID})
	err = res.One(&u)
	if err != nil {
		return nil, err
	}

	// add token to the user
	u.Token = theToken

	return &u, nil
}

// GetTokensForUser gets all tokens for a give user
func (t *Token) GetTokensForUser(id int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": id})

	err := res.All(&tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Get gets the token associated with the given id
func (t *Token) Get(id int) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": id})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// GetByToken gets a token associated with the plaintext token
func (t *Token) GetByToken(plainText string) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Delete deletes a token by the token's id
func (t *Token) Delete(id int) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}

	return nil
}

// DeleteByToken deletes the token by the token value
func (t *Token) DeleteByToken(plainText string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.Delete()
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new token associated with a given user
func (t *Token) Insert(token Token, u User) error {
	collection := upper.Collection(t.Table())

	// find and delete existing tokens associated with given user
	res := collection.Find(up.Cond{"user_id": u.ID})
	err := res.Delete()
	if err != nil {
		return err
	}

	// verify all fields are set on the token variable
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	token.FirstName = u.FirstName
	token.Email = u.Email

	// insert the new token
	_, err = collection.Insert(token)
	if err != nil {
		return err
	}

	return nil
}

// GenerateToken generates a token for a user with a time to live duration
func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID: userID,
		// add duration to current date and time
		Expires: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)
	// read into randomBytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// populate the plainText token
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// get hash
	hash := sha256.Sum256([]byte(token.PlainText))

	// populate token with the hash
	token.Hash = hash[:]

	return token, nil
}

// AuthenticateToken authenticates a token
func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get(authorization)
	// if auth header doesn't exist
	if authorizationHeader == "" {
		return nil, errNoAuthHeader
	}

	// check that header is in correct format
	headerParts := strings.Split(authorizationHeader, " ")

	// check for bearer + token
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errNoAuthHeader
	}

	token := headerParts[1]

	// check that the token is in the correct format
	if len(token) != 26 {
		return nil, errTokenWrongSize
	}

	// get token from db
	tkn, err := t.GetByToken(token)
	if err != nil {
		return nil, errTokenNoMatch
	}

	// check if token expired
	if tkn.Expires.Before(time.Now()) {
		return nil, errTokenExpired
	}

	// get user associated with token
	user, err := tkn.GetUserForToken(token)
	if err != nil {
		return nil, errUserNoMatch
	}

	return user, nil
}

// check the token is valid
func (t *Token) ValidToken(token string) (bool, error) {
	// get user associated with token
	user, err := t.GetUserForToken(token)
	if err != nil {
		return false, errUserNoMatch
	}

	// check token is not empty (eg user with no token)
	if user.Token.PlainText == "" {
		return false, errTokenNoMatch
	}

	// check if token expired
	if t.Expires.Before(time.Now()) {
		return false, errTokenExpired
	}

	// is valid non-expired token
	return true, nil
}
