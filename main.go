package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/NEUP-Net-Depart/email-daemon/server"
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"net/rpc/jsonrpc"
	"net/http"
//	"io/ioutil"
	"bytes"
	"encoding/json"
	"crypto/md5"
	"strconv"
	"encoding/hex"
	"io/ioutil"
)

var mu = sync.Mutex{}
var globCfg = config.MessageConfig{}
var MailCnt = 0

const tpl = `
亲爱的同学 %s 您好～ 有人在先锋市场给你发送了私信消息哦，点击下面链接回复吧:
https://market.neupioneer.com/message
`

func main() {
	// First init config
	_, err := toml.DecodeFile("config.toml", &config.GlobCfg)
	if err != nil {
		panic(err)
	}
	_, err = toml.DecodeFile("config.toml", &config.MessageCfg)
	if err != nil {
		panic(err)
	}
	globCfg = config.MessageCfg
	log.Infof("Config init done")

	db, err := gorm.Open("mysql", globCfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Database init done")
	defer db.Close()

	// Start HTTP Server in seperated goroutine
	log.Infof("Initialize http server")
	log.Infof("Initialize rpc server")
	log.Infof("Initialize worker")
	go server.HTTPServer()
	go server.ServeRPC()
	go server.Worker()

	for {
		// Here first traverse the whole list
		userList, err := ListUser(db)
		if err != nil {
			err = errors.Wrap(err, "main routine")
			log.Error(err)
			continue
		}
		for _, user := range userList {
			// This user has left message page, ready to receive message.
			if (time.Now().Unix()-int64(user.LastGetNewMessageTime)) > globCfg.TimeLimit {
				// He wants a new E-mail (as well as a new text)
				if user.LastGetNewMessageTime >= user.LastSendEmailTime {
					lst, err := MessagesByUserID(db, user.ID)
					if err == nil {
						if len(lst) != 0 {
							// we need to send mail
							cfg := SendConfig{}
							cfg.FromName = globCfg.FromName
							cfg.SendID = globCfg.SendID
							cfg.To = user.Email
							cfg.Body = fmt.Sprintf(tpl, user.Nickname)
							cfg.Title = globCfg.Title
							// multi-goroutine
							if !(MailCnt > 5 || user.Email == "") {
								log.Infof("Preparing to send email to user %s[%d] e-mail: %s", user.Nickname, user.ID, user.Email)
								mu.Lock()
								MailCnt++
								mu.Unlock()
								go sendMail(cfg, user, db)
							}
						}
					} else {
						err = errors.Wrap(err, "main routine")
						log.Error(err)
					}
				}
				// He wants a new Wechat
				if user.WechatOpenID != "" {
					lst, err := WeChatMessagesByUserID(db, user.ID)
					if err == nil {
						if len(lst) != 0 {
							// We need to send wechat
							for _, msg := range lst {
								sendWechat(msg, user.WechatOpenID, db)
							}
						}
					}
				}

			}

		}
		time.Sleep(time.Second * globCfg.Interval)
	}
}

// goroutine to run the mail sending fun
func sendMail(cfg SendConfig, user User, db *gorm.DB) {
	ID := user.ID
	var reply int
	var args config.MailSettings
	cli, err := jsonrpc.Dial("tcp", "127.0.0.1:65525")
	if err != nil {
		log.Error(err)
		return
	}
	args.Body = cfg.Body
	args.SendID = cfg.SendID
	args.To = cfg.To
	args.Subject = cfg.Title
	args.FromName = cfg.FromName
	err = cli.Call("Daemon.SendMail", &args, &reply)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(reply)
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

func sendWechat(msg Message, openID string, db *gorm.DB) {

	log.Infof("Sending wx message [%d] from user [%d] to user [%d] DONE", msg.ID, msg.SenderID, msg.ReceiverID)

	var sender_name string
	if msg.SenderID == 0 {
		sender_name = "系统消息"
	} else {
		sender_name = msg.Sender.Nickname
	}

	t := time.Now().Unix()

	datas := [5]map[string]string {
		{ "name": "first", "value": "【先锋市场】新消息提醒" },
		{ "name": "keyword1", "value": sender_name },
		{ "name": "keyword2", "value": msg.Content },
		{ "name": "keyword3", "value": msg.CreatedAt.Format("2006-01-02 15:04:05") },
		{ "name": "remark", "value": "您收到一条系统消息，请及时查看。" },
	}

	data := map[string]interface{} {
		"toUser": openID,
		"templateId": "knlItrLhqCnJNIzQRntDIXggv4tpJJ0U_ODbm3kPIcc",
		"url": "/message",
		"datas": datas,
	}

	data_json, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return
	}
	data_str := string(data_json)
	biz := "market.neupioneer"

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(config.GlobCfg.WechatMsgKey + biz + data_str + strconv.FormatInt(t,10)))
	sign := hex.EncodeToString(md5Ctx.Sum(nil))

	xdata := map[string]interface{} {
		"timestamp": t,
		"data": data_str,
		"bizCode": biz,
		"sign": sign,
	}

	xdata_json, err := json.Marshal(xdata)
	if err != nil {
		log.Error(err)
		return
	}

	client := &http.Client{}
	req_buf := bytes.NewBuffer(xdata_json)
	request, _ := http.NewRequest("POST", "https://api.xms.rmbz.net/open/msg/send", req_buf)
	request.Header.Set("Content-type", "application/json")
	response, _ := client.Do(request)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		log.Info(string(body))
		err = SetMsgWechatSent(db, &msg)
		if err != nil {
			log.Error(err)
		}
		log.Infof("Sent wx message [%d] to wechat [%s]", msg.ID, openID)
	}
}