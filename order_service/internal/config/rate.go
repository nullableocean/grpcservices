package config

import "time"

type RolesLimitConfig struct {
	GuestMaxRequests       int           `env:"RATE_LIMIT_GUEST_MAX" env-default:"10"`
	GuestWindow            time.Duration `env:"RATE_LIMIT_GUEST_WINDOW" env-default:"1m"`
	TraderMaxRequests      int           `env:"RATE_LIMIT_TRADER_MAX" env-default:"100"`
	TraderWindow           time.Duration `env:"RATE_LIMIT_TRADER_WINDOW" env-default:"1m"`
	MarketMakerMaxRequests int           `env:"RATE_LIMIT_MM_MAX" env-default:"1000"`
	MarketMakerWindow      time.Duration `env:"RATE_LIMIT_MM_WINDOW" env-default:"1m"`
	ModerMaxRequests       int           `env:"RATE_LIMIT_MODER_MAX" env-default:"500"`
	ModerWindow            time.Duration `env:"RATE_LIMIT_MODER_WINDOW" env-default:"1m"`
	AdminMaxRequests       int           `env:"RATE_LIMIT_ADMIN_MAX" env-default:"10000"`
	AdminWindow            time.Duration `env:"RATE_LIMIT_ADMIN_WINDOW" env-default:"1m"`
}
