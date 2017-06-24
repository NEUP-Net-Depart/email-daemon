package server

import (
	"net/rpc/jsonrpc"
	"testing"

	"github.com/NEUP-Net-Depart/email-daemon/config"
)

func TestServeRPC(t *testing.T) {
	// go ServeRPC()
}

func TestRPCSendMail(t *testing.T) {
	var reply int
	var args config.MailSettings
	cli, err := jsonrpc.Dial("tcp", "127.0.0.1:65525")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	args.Body = "<h1>我要测试</h1>"
	args.SendID = "notify"
	args.To = "zhangjianqiu_133@yeah.net"
	args.Subject = "邮件测试哦OwO"
	args.FromName = "我"
	err = cli.Call("Daemon.SendMail", &args, &reply)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Log(reply)
}
