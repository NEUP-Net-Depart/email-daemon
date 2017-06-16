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

var MailSentMap map[int]int
var mailSendingMap map[int]int
var mu = sync.Mutex{}
var globCfg = Config{}

const tpl = `
亲爱的同学 %s 您好～ 有人在先锋市场给你发送了私信消息哦，点击下面链接回复吧:
<h3><a href=https://market.neupioneer.com/message>戳我戳我</a></h3>
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
	MailSentMap = make(map[int]int)
	mailSendingMap = make(map[int]int)

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
			if MailSentMap[user.ID] == 1 || mailSendingMap[user.ID] == 1 {
				// This user already sent, continue
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
				log.Infof("Preparing to send email to user %s[%d] e-mail: %s", user.Nickname, user.ID, user.Email)
				go sendmail(cfg, user.ID)
			}
		}
		time.Sleep(time.Hour * globCfg.Interval)
	}
}

// goroutine to run the mail sending fun
func sendmail(cfg SendConfig, ID int) {
	mu.Lock()
	mailSendingMap[ID] = 1
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
		mailSendingMap[ID] = 0
		return
	}
	// Else update the send status
	mu.Lock()
	MailSentMap[ID] = 1
	mailSendingMap[ID] = 0
	mu.Unlock()
	log.Infof("Sent email to user [%d] DONE", ID)
}
