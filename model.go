package main

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type User struct {
	ID       int
	Username string
	Nickname string
	Email    string
}

type Message struct {
	ReceiverID int
	IsRead     bool
}

type SendConfig struct {
	From     string
	To       string
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	Body     string
	Title    string
}

func ListUser(db *gorm.DB) (users []User, err error) {
	err = db.Table("users").Scan(&users).Error
	if err != nil {
		err = errors.Wrap(err, "ListUser")
		return
	}
	return
}

func MessagesByUserID(db *gorm.DB, userID int) (msg []Message, err error) {
	err = db.Table("message").Where("receiver_id = ?", userID).Where("is_read = ?", false).Scan(&msg).Error
	if err != nil {
		err = errors.Wrap(err, "MessagesByUserID")
		return
	}
	return
}
