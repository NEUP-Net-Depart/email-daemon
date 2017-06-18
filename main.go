package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

var mu = sync.Mutex{}
var globCfg = Config{}
var MailCnt = 0

const tpl = `
亲爱的同学 %s 您好～ 有人在先锋市场给你发送了私信消息哦，点击下面链接回复吧:
https://market.neupioneer.com/message
`

func main() {
	// First init config
	_, err := toml.DecodeFile("config.toml", &globCfg)
	if err != nil {
		panic(err)
	}
	log.Infof("Config init done")

	db, err := gorm.Open("mysql", globCfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Database init done")
	defer db.Close()

	// Start HTTP Server in seperated goroutine
	go HTTPServer()
	log.Infof("Http Server init done")

	for {
		// Here first traverse the whole list
		userList, err := ListUser(db)
		if err != nil {
			err = errors.Wrap(err, "main routine")
			log.Error(err)
			continue
		}
		for _, user := range userList {
			if !(user.LastGetNewMessageTime >= user.LastSendEmailTime &&
				(time.Now().Unix() - int64(user.LastGetNewMessageTime)) > globCfg.TimeLimit) {
				// This user have already sent
				continue
			}
			lst, err := MessagesByUserID(db, user.ID)
			if err != nil {
				err = errors.Wrap(err, "main routine")
				log.Error(err)
				continue
			}
			if len(lst) != 0 {
				// we need to send mail
				cfg := SendConfig{}
				cfg.From = globCfg.FromAddress
				cfg.SMTPHost = globCfg.SMTPHost
				cfg.SMTPPass = globCfg.SMTPPass
				cfg.SMTPUser = globCfg.SMTPUser
				cfg.SMTPPort = globCfg.SMTPPort
				cfg.To = user.Email
				cfg.Body = fmt.Sprintf(tpl, user.Nickname)
				cfg.Title = globCfg.Title
				// multi-goroutine
				if MailCnt > 5 || user.Email == "" {
					continue
				}
				log.Infof("Preparing to send email to user %s[%d] e-mail: %s", user.Nickname, user.ID, user.Email)
				mu.Lock()
				MailCnt++
				mu.Unlock()
				go sendmail(cfg, user, db)
			}
		}
		time.Sleep(time.Hour * globCfg.Interval)
	}
}

// goroutine to run the mail sending fun
func sendmail(cfg SendConfig, user User, db *gorm.DB) {
	ID := user.ID
	mu.Lock()
	mu.Unlock()
	cli := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	cli.SSL = true
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", cfg.To)
	m.SetHeader("Subject", cfg.Title)
	m.SetBody("text/html", cfg.Body)
	err := cli.DialAndSend(m)
	if err != nil {
		log.Error(err)
		return
	}
	// Else update the send status
	mu.Lock()
	err = SetUserEmailLock(db, &user)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("Updated user [%d] lastSendEmailTime", ID)
	MailCnt--
	mu.Unlock()
	log.Infof("Sent email to user [%d] DONE", ID)
}
