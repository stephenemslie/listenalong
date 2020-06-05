package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/rs/xid"
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

type UserService struct {
	db        *dynamo.DB
	userTable dynamo.Table
}

func (u *UserService) CreateTable() {
	err := u.db.CreateTable("users", User{}).Run()
	if err != nil {
		fmt.Println("error", err)
	}
}

func (u *UserService) CreateUser(user *User) error {
	user.Id = xid.New().String()
	return u.userTable.Put(user).Run()
}

func (u *UserService) GetUser(user *User) error {
	err := u.userTable.Get("user_id", user.Id).One(user)
	if err != nil {
		fmt.Println("get error", err)
		return err
	}
	return nil
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
