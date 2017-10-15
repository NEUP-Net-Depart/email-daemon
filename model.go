package main

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"time"
)

type User struct {
	ID                    int
	Username              string
	Nickname              string
	Email                 string
	Tel                   string
	LastGetNewMessageTime int
	LastSendEmailTime     int
	LastSendWxTime        int
	WechatOpenID          string
}

type Message struct {
	ID         int
	ReceiverID int
	IsRead     bool
	WxSent     bool
	Content    string
	SenderID   int
	Sender     User `gorm:"ForeignKey:SenderID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SendConfig struct {
	FromName string
	SendID   string
	To       string
	Body     string
	Title    string
}

type TextConfig struct {
	To   string
	Body string
}

func (Message) TableName() string {
	return "message"
}

func ListUser(db *gorm.DB) (users []User, err error) {
	err = db.Find(&users).Error
	if err != nil {
		err = errors.Wrap(err, "ListUser")
		return
	}
	return
}

func MessagesByUserID(db *gorm.DB, userID int) (msgs []Message, err error) {
	err = db.Where("receiver_id = ?", userID).Where("is_read = ?", false).Find(&msgs).Error
	if err != nil {
		err = errors.Wrap(err, "MessagesByUserID")
		return
	}
	return
}

func WeChatMessagesByUserID(db *gorm.DB, userID int) (msgs []Message, err error) {
	err = db.Where("receiver_id = ?", userID).Where("is_read = ?", false).Where("wx_sent = ?", false).Preload("Sender").Find(&msgs).Error
	if err != nil {
		err = errors.Wrap(err, "MessagesByUserID")
		return
	}
	return
}

func SetUserEmailLock(db *gorm.DB, user *User) (err error) {
	err = db.Model(user).Where("id = ?", user.ID).Update("last_send_email_time", int(time.Now().Unix())).Error
	if err != nil {
		err = errors.Wrap(err, "SetUserEmailLock")
		return
	}
	return
}

func SetUserWxLock(db *gorm.DB, user *User) (err error) {
	err = db.Model(user).Where("id = ?", user.ID).Update("last_send_wx_time", int(time.Now().Unix())).Error
	if err != nil {
		err = errors.Wrap(err, "SetUserWxLock")
		return
	}
	return
}

func SetMsgWechatSent(db *gorm.DB, msg *Message) (err error) {
	err = db.Table("message").Where("id = ?", msg.ID).Updates(map[string]interface{}{"wx_sent": true}).Error
	if err != nil {
		err = errors.Wrap(err, "SetMsgWechatSent")
		return
	}
	return
}
