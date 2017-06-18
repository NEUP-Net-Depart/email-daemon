package main

import (
	"testing"

	"github.com/jinzhu/gorm"
)

func TestSendMail(t *testing.T) {
	db, err := gorm.Open("mysql", "root:@/flea?charset=utf8")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	user := User{}
	user.ID = 1
	cfg := SendConfig{}
	cfg.From = "notification@neup.market"
	cfg.SMTPHost = "smtp.yandex.com"
	cfg.SMTPPort = 465
	cfg.SMTPUser = "notification@neup.market"
	cfg.SMTPPass = "fleamarket@neup"
	cfg.Title = "邮件发送测试"
	cfg.Body = `
			<h1>测试一下</h1>
			这是一封测试邮件
	`
	// cfg.To = "lijiahao@cool2645.com"
	cfg.To = "zhangjianqiu_133@yeah.net"
	sendmail(cfg, user, db)
}

func TestGetUserList(t *testing.T) {
	db, err := gorm.Open("mysql", "root:@/flea?charset=utf8")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	lst, err := ListUser(db)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("%+v", lst)
}

func TestGetUserMessageList(t *testing.T) {
	db, err := gorm.Open("mysql", "root:@/flea?charset=utf8")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	lst, err := MessagesByUserID(db, 1)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("%+v", lst)
}

func TestSetUserEmailLock(t *testing.T) {
	db, err := gorm.Open("mysql", "root:@/flea?charset=utf8")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	user := User{}
	user.ID = 1
	err = SetUserEmailLock(db, &user)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("Updated user [%d] lastSendEmailTime", user.ID)
}
