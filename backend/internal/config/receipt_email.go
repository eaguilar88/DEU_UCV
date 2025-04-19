package config

type ReceiptEmail struct {
	Addr     string `env:"RECEIPT_EMAIL_ADDR" envDefault:"receipt-email:8082" envWhitelisted:"true"`
	PingAddr string `env:"RECEIPT_EMAIL_PING_ADDR" envDefault:"http://receipt-email/heartbeat" envWhitelisted:"true"`
}
