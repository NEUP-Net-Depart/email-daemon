package server

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/BurntSushi/toml"
	"github.com/NEUP-Net-Depart/email-daemon/config"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type ServerList struct {
	Servers map[string]config.Config
}

var sl ServerList

type Daemon int

// SendMail is a RPC service that serves the call
func (d *Daemon) SendMail(ms *config.MailSettings, reply *int) (err error) {
	log.Infof("Send Mail with settings %+v", ms)
	toml.DecodeFile("server.toml", &sl)
	if _, ok := sl.Servers[ms.SendID]; !ok {
		err = errors.New("Invaild SendID, please check the config file")
		return
	}
	return
}

func dispatcher() {

}

func ServeRPC() {
	ms := new(Daemon)
	rpc.Register(ms)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":65525")
	if err != nil {
		log.Fatalf("Cannot start RPC service: %s", err)
	}
	http.Serve(l, nil)
}
