package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/rs/xid"
)

type User struct {
	UserID string `dynamo:"user_id,hash"`
	Name   string `dynamo:"name"`
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
	user.ID = xid.New().String()
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
