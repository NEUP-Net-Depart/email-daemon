package server

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/NEUP-Net-Depart/email-daemon/mq"
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
	log.Infof("RPC Send Mail with settings %+v", ms)
	toml.DecodeFile("server.toml", &sl)
	if _, ok := sl.Servers[ms.SendID]; !ok {
		err = errors.New("Invaild SendID, please check the config file")
		return
	}
	server := sl.Servers[ms.SendID]
	// Set up the job then push into job queue
	j := mq.Job{}
	j.Body = ms.Body
	j.To = ms.To
	j.From = server.FromAddress
	j.Title = ms.Subject
	j.TimeStamp = time.Now()
	j.SMTPHost = server.SMTPHost
	j.SMTPUser = server.SMTPUser
	j.SMTPPass = server.SMTPPass
	j.SMTPPort = server.SMTPPort
	err = mq.PushJob(j)
	if err != nil {
		err = errors.Wrap(err, "Daemon.SendMail")
		return
	}
	log.Infof("Job pushed to queue %+v", j)
	return
}

func Worker() {
	mq.Init()
	log.Infof("RPC Mail worker runnning ...")
	for {
		j, err := mq.GetJob()
		if err != nil {
			log.Errorf("worker: %s", err)
		}
		err = j.Send()
		if err != nil {
			log.Errorf("worker: %s", err)
		}
	}
	mq.Destroy()
}

func ServeRPC() {
	ms := new(Daemon)
	rpc.Register(ms)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:65525")
	if err != nil {
		log.Fatalf("ServeRPC: Cannot start RPC service: %s", err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorf("ServeRPC: %s", err)
		}
		go jsonrpc.ServeConn(conn)
	}
}
