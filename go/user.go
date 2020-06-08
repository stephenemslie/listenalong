package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"io/ioutil"
	"net/http"
)

type SpotifyContext struct {
	URI  string `json:"uri"`
	Type string `json:"type"`
}

type SpotifyItem struct {
	ID       string `json:"id"`
	URI      string `json:"uri"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Duration int    `json:"duration_ms"`
}

type SpotifyPlaying struct {
	Context   *SpotifyContext `json:"context"`
	Item      *SpotifyItem    `json:"item"`
	Progress  int             `json:"progress_ms"`
	IsPlaying bool            `json:"is_playing"`
}

type User struct {
	ID           string    `dynamo:"id,hash" json:"id"`
	Following    string    `dynamo:"following_id"`
	Name         string    `dynamo:"name" json:"display_name"`
	CreatedAt    time.Time `dynamo:"created_at"`
	UpdatedAt    time.Time `dynamo:"updated_at"`
	Playing      bool      `dynamo:"is_playing"`
	Progress     int       `dynamo:"progress_ms"`
	ContextURI   string    `dynamo:"context_uri"`
	ContextType  string    `dynamo:"context_type"`
	ItemID       string    `dynamo:"item_id"`
	ItemURI      string    `dynamo:"item_uri"`
	ItemName     string    `dynamo:"item_name"`
	ItemDuration int       `synamo:"item_duration"`
}

func NewUserFromSpotify(client *http.Client) (User, error) {
	res, err := client.Get("https://api.spotify.com/v1/me")
	user := User{}
	if err != nil {
		return user, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (u *User) UpdatePlaying(client *http.Client) (bool, error) {
	res, err := client.Get("https://api.spotify.com/v1/me/player/currently-playing")
	if err != nil {
		return false, err
	}
	if res.StatusCode == 204 {
		return false, nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	spotifyPlaying := SpotifyPlaying{}
	err = json.Unmarshal(body, &spotifyPlaying)
	if err != nil {
		return false, err
	}
	now := time.Now()
	u.UpdatedAt = now
	u.Playing = spotifyPlaying.IsPlaying
	u.Progress = spotifyPlaying.Progress
	u.ContextURI = spotifyPlaying.Context.URI
	u.ContextType = spotifyPlaying.Context.Type
	u.ItemID = spotifyPlaying.Item.ID
	u.ItemURI = spotifyPlaying.Item.URI
	u.ItemName = spotifyPlaying.Item.Name
	u.ItemDuration = spotifyPlaying.Item.Duration
	return true, nil
}

type UserService struct {
	db        *dynamo.DB
	userTable dynamo.Table
}

func (u *UserService) CreateTable() error {
	err := u.db.CreateTable("users", User{}).Run()
	if err != nil {
		fmt.Println("error", err)
		return err
	}
	return nil
}

func (u *UserService) CreateUser(user *User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	return u.userTable.Put(user).Run()
}

func (u *UserService) GetUser(userID string, user *User) error {
	err := u.userTable.Get("id", userID).One(user)
	if err != nil {
		fmt.Println("get error", err)
		return err
	}
	return nil
}

func (u *UserService) GetOrCreateUser(user *User) error {
	err := u.GetUser(user.ID, user)
	if err != dynamo.ErrNotFound {
		return err
	}
	err = u.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserService) PutUser(user *User) error {
	return u.userTable.Put(user).Run()
}

func NewUserService(endpoint string) (*UserService, error) {
	db, dynamoTable, err := newDynamoTable("users", endpoint)
	if err != nil {
		return nil, err
	}
	userService := &UserService{
		db:        db,
		userTable: dynamoTable,
	}
	return userService, err
}

func newDynamoTable(tableName, endpoint string) (*dynamo.DB, dynamo.Table, error) {
	cfg := aws.Config{
		Region: aws.String("eu-west-1"),
	}
	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}
	sess := session.Must(session.NewSession())
	db := dynamo.New(sess, &cfg)
	table := db.Table(tableName)
	return db, table, nil
}
