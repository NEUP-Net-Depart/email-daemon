package server

import (
	"net/rpc"
	"testing"
	"time"

	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/NEUP-Net-Depart/email-daemon/mq"
)

func TestServeRPC(t *testing.T) {
	// go ServeRPC()
}

func TestRPCSendMail(t *testing.T) {
	go ServeRPC()
	time.Sleep(time.Second * 5)
	var reply int
	var args config.MailSettings
	cli, err := rpc.DialHTTP("tcp", "127.0.0.1:65525")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	err = cli.Call("Daemon.SendMail", &args, &reply)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
