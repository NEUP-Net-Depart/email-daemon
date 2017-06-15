package main

type Config struct {
	SMTPPort    int    `toml:"smtp_port"`
	SMTPPass    string `toml:"smtp_pass"`
	SMTPUser    string `toml:"smtp_user"`
	SMTPHost    string `toml:"smtp_host"`
	FromAddress string `toml:"from_address"`
	Title       string `toml:"title"`
	DSN         string `toml:"dsn"`
}
