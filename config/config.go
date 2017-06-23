package config

import "time"

type Config struct {
	SMTPPort    int           `toml:"smtp_port"`
	SMTPPass    string        `toml:"smtp_pass"`
	SMTPUser    string        `toml:"smtp_user"`
	SMTPHost    string        `toml:"smtp_host"`
	FromAddress string        `toml:"from_address"`
	Title       string        `toml:"title"`
	DSN         string        `toml:"dsn"`
	Interval    time.Duration `toml:"interval"`
	TimeLimit   int64         `toml:"time_limit"`
}

type MailSettings struct {
	To      string
	SendID  string
	Subject string
	Body    string
}
