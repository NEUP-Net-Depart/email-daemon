package config

import "time"

var GlobCfg = Config{}
var MessageCfg = MessageConfig{}

type Config struct {
	AMQPConfig  	string        `toml:"amqp_config"`
	WechatMsgKey	string		  `toml:"wechat_msg_key"`
	TextApiKey		string        `toml:"text_api_key"`
	Debug			bool		  `toml:"debug"`
}

type SMTPConfig struct {
	SMTPPort    int           `toml:"smtp_port"`
	SMTPPass    string        `toml:"smtp_pass"`
	SMTPUser    string        `toml:"smtp_user"`
	SMTPHost    string        `toml:"smtp_host"`
	FromAddress string        `toml:"from_address"`
}

type MailSettings struct {
	To		    string
	FromName    string
	SendID      string
	Subject     string
	Body        string
}

type MessageConfig struct {
	Title       string        `toml:"title"`
	SendID      string        `toml:"send_id"`
	FromName    string        `toml:"from_name"`
	DSN         string        `toml:"dsn"`
	Interval    time.Duration `toml:"interval"`
	TimeLimit   int64         `toml:"time_limit"`
}