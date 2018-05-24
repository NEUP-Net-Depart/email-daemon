package main

import (
	"fmt"
	"time"

	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/NEUP-Net-Depart/email-daemon/server"
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
)

var globCfg = config.MessageConfig{}

const (
	tpl = `
亲爱的同学 %s 您好～ 有人在先锋市场给你发送了私信消息哦，点击下面链接回复吧:
https://market.neupioneer.com/message
`
	tplTel = "【先锋市场】亲，您有%d条未读消息，请及时处理，点击 https://market.neupioneer.com/message 来查看。"
)

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

	if config.GlobCfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

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
			// This user has left message page, ready to receive email and text.
			// He has new unknown messages
			if (time.Now().Unix()-int64(user.LastGetNewMessageTime)) > globCfg.TimeLimit &&
				(user.LastGetNewMessageTime >= user.LastSendEmailTime ||
					int(time.Now().Unix())-user.LastSendEmailTime > int(globCfg.EmailLockTime)*60) {

				var s = "User [%d] has read his message and left message page, "
				if user.LastGetNewMessageTime >= user.LastSendEmailTime {
					s += "he has never received a notification since last visit message page, "
				} else {
					s += "he received a notification last time one day ago, "
				}
				log.Debugf(s+"is willing to receive a notification.", user.ID)

				// Send Email & Text
				lst, err := MessagesByUserID(db, user.ID)
				if err == nil {
					if len(lst) != 0 {
						// Set notified state
						// all notifications will be blocked until his visited message page or a day passed
						err = SetUserEmailLock(db, &user)
						if err != nil {
							log.Error(err)
							return
						}
						log.Infof("Updated user [%d] lastSendEmailTime", user.ID)

						// we need to send mail
						cfg := SendConfig{}
						cfg.FromName = globCfg.FromName
						cfg.SendID = globCfg.SendID
						cfg.To = user.Email
						cfg.Body = fmt.Sprintf(tpl, user.Nickname)
						cfg.Title = globCfg.Title
						// multi-goroutine
						if !(user.Email == "") {
							log.Infof("Preparing to send email to user %s[%d] e-mail: %s", user.Nickname, user.ID, user.Email)
							go sendMail(cfg, user, db)
						}

						// we need to send text
						tcfg := TextConfig{}
						tcfg.To = user.Tel
						tcfg.Body = fmt.Sprintf(tplTel, len(lst))
						// multi-goroutine
						if !(user.Tel == "") {
							log.Infof("Preparing to send text to user %s[%d] tel: %s", user.Nickname, user.ID, user.Tel)
							go sendText(tcfg, user, db)
						}
					}
				} else {
					err = errors.Wrap(err, "main routine")
					log.Error(err)
				}
			} else {
				var s = "User [%d] "
				if !(time.Now().Unix()-int64(user.LastGetNewMessageTime) > globCfg.TimeLimit) {
					s += "is likely to be on message page, "
				} else {
					s += "received a notification in one day, "
				}
				log.Debugf(s+"is NOT willing to receive a notification.", user.ID)
			}

			// This user has left message page, ready to receive wx.
			// He has new unknown messages
			if (time.Now().Unix()-int64(user.LastGetNewMessageTime)) > globCfg.TimeLimit &&
				(user.LastGetNewMessageTime >= user.LastSendWxTime ||
					int(time.Now().Unix())-user.LastSendWxTime > int(globCfg.WxLockTime)*60) {

				var s = "User [%d] has read his message and left message page, "
				if user.LastGetNewMessageTime >= user.LastSendWxTime {
					s += "he has never received a wx since last visit message page, "
				} else {
					s += "he received a wx last time 10 minutes ago, "
				}
				log.Debugf(s+"is willing to receive a notification.", user.ID)
				// Send wechat
				if user.WechatOpenID != "" {
					lst, err := WeChatMessagesByUserID(db, user.ID)
					if err == nil {
						if len(lst) != 0 {
							// Set notified state
							// all wx will be blocked until his visited message page or 10 minutes passed
							err = SetUserWxLock(db, &user)
							if err != nil {
								log.Error(err)
								return
							}
							log.Infof("Updated user [%d] lastSendWxTime", user.ID)
							go sendWechat(lst, user.WechatOpenID, db, len(lst))
						}
					}
				}
			} else {
				var s = "User [%d] "
				if !(time.Now().Unix()-int64(user.LastGetNewMessageTime) > globCfg.TimeLimit) {
					s += "is likely to be on message page, "
				} else {
					s += "received a wx in 10 minutes, "
				}
				log.Debugf(s+"is NOT willing to receive a wx.", user.ID)
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
	log.Infof("Sent email to user [%d] DONE", ID)
}

// goroutine to run the mail sending fun
func sendText(cfg TextConfig, user User, db *gorm.DB) {
	ID := user.ID
	data := strings.NewReader("apikey=" + config.GlobCfg.TextApiKey + "&mobile=" + cfg.To + "&text=" + cfg.Body)
	client := &http.Client{}
	request, err := http.NewRequest("POST", "https://sms.yunpian.com/v2/sms/single_send.json", data)
	if err != nil {
		log.Error(err)
		return
	}
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return
	}
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		log.Info(string(body))
		log.Infof("Sent text to user [%d] DONE", ID)
	}
}

func sendWechat(msgs []Message, openID string, db *gorm.DB, num int) {

	var msg = msgs[0]
	var last_msg = msgs[len(msgs)-1]
	if len(msgs) == 1 {
		log.Infof("Sending wx message [%d] from user [%d] to user [%d]", msg.ID, msg.SenderID, msg.ReceiverID)
	} else {
		log.Infof("Sending wx message [%d-%d] from user [%d] to user [%d]", msg.ID, last_msg.ID,
			msg.SenderID, msg.ReceiverID)
	}

	var sender_name string
	if msg.SenderID == 0 {
		sender_name = "系统消息"
	} else {
		sender_name = msg.Sender.Nickname
	}

	t := time.Now().Unix() // 请务必确认服务器时间是准确的

	var datas [5]map[string]string
	if num == 1 {
		datas = [5]map[string]string{
			{"name": "first", "value": "【先锋市场】新消息提醒"},
			{"name": "keyword1", "value": sender_name},
			{"name": "keyword2", "value": msg.Content},
			{"name": "keyword3", "value": msg.CreatedAt.Format("2006-01-02 15:04:05")},
			{"name": "remark", "value": "您收到一条新消息，请及时查看。"},
		}
	} else {
		datas = [5]map[string]string{
			{"name": "first", "value": "【先锋市场】新消息提醒"},
			{"name": "keyword1", "value": sender_name + " 等"},
			{"name": "keyword2", "value": msg.Content},
			{"name": "keyword3", "value": msg.CreatedAt.Format("2006-01-02 15:04:05")},
			{"name": "remark", "value": "您收到 " + strconv.Itoa(num) + " 条新消息，请及时查看。"},
		}
	}
	log.Info(datas)

	data := map[string]interface{}{
		"toUser":     openID,
		"templateId": "knlItrLhqCnJNIzQRntDIXggv4tpJJ0U_ODbm3kPIcc",
		"url":        "/message",
		"datas":      datas,
	}

	data_json, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return
	}
	data_str := string(data_json)
	biz := "market.neupioneer"

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(config.GlobCfg.WechatMsgKey + biz + data_str + strconv.FormatInt(t, 10)))
	sign := hex.EncodeToString(md5Ctx.Sum(nil))

	xdata := map[string]interface{}{
		"timestamp": t,
		"data":      data_str,
		"bizCode":   biz,
		"sign":      sign,
	}

	xdata_json, err := json.Marshal(xdata)
	if err != nil {
		log.Error(err)
		return
	}

	client := &http.Client{}
	req_buf := bytes.NewBuffer(xdata_json)
	request, err := http.NewRequest("POST", "https://api.xms.rmbz.net/open/msg/send", req_buf)
	if err != nil {
		log.Error(err)
		return
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return
	}
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		log.Info(string(body))
		for _, msg := range msgs {
			err = SetMsgWechatSent(db, &msg)
			if err != nil {
				log.Error(err)
			}
		}

		if len(msgs) == 1 {
			log.Infof("Sent wx message [%d] to wechat [%s] DONE", msg.ID, openID)
		} else {
			log.Infof("Sent wx message [%d-%d] to wechat [%s] DONE", msg.ID, last_msg.ID, openID)
		}
	}
}
