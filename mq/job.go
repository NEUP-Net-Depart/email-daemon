package mq

// Job is the structure to send email
type Job struct {
	From      string
	To        string
	SMTPUser  string
	SMTPPass  string
	SMTPHost  string
	SMTPPort  int
	Title     string
	Body      string
	TimeStamp int64
}
