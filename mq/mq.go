package mq

import (
	"encoding/json"
	"time"

	"github.com/NEUP-Net-Depart/email-daemon/config"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var initOK bool
var ch *amqp.Channel

func Init() (err error) {
	conn, err = amqp.Dial(config.GlobCfg.AMQPConfig)
	if err != nil {
		err = errors.Wrap(err, "mq init")
		return
	}
	initOK = true
	return
}

// Return the job queue as a channel to caller
func GetJob() (j Job, err error) {
	if !initOK {
		err = errors.New("mq should init before use")
		return
	}
	// Open a channel for connection
	ch, err = conn.Channel()
	if err != nil {
		err = errors.Wrap(err, "mq get job")
		return
	}
	// Declare a queue
	_, er := ch.QueueDeclare("emails", false, false, false, false, nil)
	if er != nil {
		err = errors.Wrap(er, "mq get job")
		return
	}
	// get one item
	for {
		m, ok, er := ch.Get("emails", true)
		if er != nil {
			err = errors.Wrap(er, "mq get job")
			return
		}
		// If no jobs then loop here
		if ok {
			err = json.Unmarshal(m.Body, &j)
			if err != nil {
				err = errors.Wrap(err, "mq get job")
			}
			return
		}
		// do not stress CPU
		time.Sleep(time.Second * 1)
	}
	return
}

func PushJob(j Job) (err error) {
	if !initOK {
		err = errors.New("mq should init before use")
		return
	}
	ch, err = conn.Channel()
	if err != nil {
		err = errors.Wrap(err, "mq push job")
		return
	}
	_, er := ch.QueueDeclare("emails", false, false, false, false, nil)
	if er != nil {
		err = errors.Wrap(er, "mq push job")
		return
	}
	body, er := json.Marshal(j)
	if er != nil {
		err = errors.Wrap(er, "mq push job")
		return
	}
	err = ch.Publish("", "emails", true, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
	if err != nil {
		err = errors.Wrap(err, "mq push job")
		return
	}
	return
}

func Destroy() {
	if !initOK {
		panic("mq should init before destroy")
	}
	conn.Close()
}
