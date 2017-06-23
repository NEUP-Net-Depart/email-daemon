package mq

import (
	"testing"
	"time"

	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/streadway/amqp"
)

func TestConn(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	conn.Close()
}

func TestInit(t *testing.T) {
	config.GlobCfg.AMQPConfig = "amqp://guest:guest@localhost:5672/"
	err := Init()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return
}

func TestPushJob(t *testing.T) {
	config.GlobCfg.AMQPConfig = "amqp://guest:guest@localhost:5672/"
	err := Init()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	j := Job{}
	j.Body = "This is a body"
	j.From = "test@example.com"
	j.To = "to@example.com"
	j.Title = "Email"
	j.SMTPHost = "example.com"
	j.SMTPPort = 465
	j.SMTPPass = "password"
	j.SMTPUser = "user@example.com"
	j.TimeStamp = time.Now()
	err = PushJob(j)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return
}

func TestGetJob(t *testing.T) {
	config.GlobCfg.AMQPConfig = "amqp://guest:guest@localhost:5672/"
	err := Init()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("Init done")
	j, err := GetJob()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("%+v", j)
}
