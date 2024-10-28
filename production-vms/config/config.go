package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// modeled after e2e/config/azure.go

var (
	Config            = mustLoadConfig()
	Azure             = mustNewAzureClient(Config.SubscriptionID)
	ResourceGroupName = "alisonproduction-testvm-rg"
)

type Configuration struct {
	SubscriptionID string        `env:"SUBSCRIPTION_ID" envDefault:"8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8"`
	Location       string        `env:"LOCATION" envDefault:"westus3"`
	TestTimeout    time.Duration `env:"TEST_TIMEOUT" envDefault:"10m"`
}

func mustLoadConfig() Configuration {
	_ = godotenv.Load(".env")
	cfg := Configuration{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
