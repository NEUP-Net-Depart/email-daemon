package mq

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	gomail "gopkg.in/gomail.v2"
)

// Job is the structure to send email
type Job struct {
	From      string    `json:"from"`
	FromName  string    `json:"from_name"`
	To        string    `json:"to"`
	SMTPUser  string    `json:"smtp_user"`
	SMTPPass  string    `json:"smtp_pass"`
	SMTPHost  string    `json:"smtp_host"`
	SMTPPort  int       `json:"smtp_port"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	TimeStamp time.Time `json:"time_stamp"`
}

// Send retry three times if we cannot send
//
func (j Job) Send() (err error) {
	maxTry := 3
	cli := gomail.NewDialer(j.SMTPHost, j.SMTPPort, j.SMTPUser, j.SMTPPass)
	cli.SSL = true
	m := gomail.NewMessage()
	m.SetHeader("To", j.To)
	m.SetAddressHeader("From", j.From, j.FromName)
	m.SetHeader("Subject", j.Title)
	m.SetBody("text/html", j.Body)
Retry:
	err = cli.DialAndSend(m)
	if err != nil {
		log.Errorf("Job.Send error: %s", err)
		maxTry--
		if maxTry == 0 {
			err = errors.Wrap(err, "Job.Send error(after 3 retries)")
			return err
		}
		goto Retry
	}
	return
}
