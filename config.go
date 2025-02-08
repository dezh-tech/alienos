package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	WorkingDirectory string   `mapstructure:"alienos_WORK_DIR"`
	RelayName        string   `mapstructure:"alienos_RELAY_NAME"`
	RelayIcon        string   `mapstructure:"alienos_RELAY_ICON"`
	RelayBanner      string   `mapstructure:"alienos_RELAY_BANNER"`
	RelayDescription string   `mapstructure:"alienos_RELAY_DESCRIPTION"`
	RelayPublicKey   string   `mapstructure:"alienos_RELAY_PUBKEY"`
	RelayContact     string   `mapstructure:"alienos_RELAY_CONTACT"`
	RelaySelf        string   `mapstructure:"alienos_RELAY_SELF"`
	RelayPort        string   `mapstructure:"alienos_RELAY_PORT"`
	RelayBind        string   `mapstructure:"alienos_RELAY_BIND"`
	RelayURL         string   `mapstructure:"alienos_RELAY_URL"`
	WhiteListed      bool     `mapstructure:"alienos_WHITE_LISTED"`
	Admins           []string `mapstructure:"alienos_ADMINS"`
}

func LoadConfig() {
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")

	viper.SetDefault("ALIENOS_WORK_DIR", "alienos_wd/")
	viper.SetDefault("ALIENOS_RELAY_NAME", "alienos")
	viper.SetDefault("ALIENOS_RELAY_ICON", "todo")
	viper.SetDefault("ALIENOS_RELAY_BANNER", "todo")
	viper.SetDefault("ALIENOS_RELAY_DESCRIPTION", "A self-hosting Nostr stack!")
	viper.SetDefault("ALIENOS_RELAY_PUBKEY", "badbdda507572b397852048ea74f2ef3ad92b1aac07c3d4e1dec174e8cdc962a")
	viper.SetDefault("ALIENOS_RELAY_CONTACT", "hi@dezh.tech")
	viper.SetDefault("ALIENOS_RELAY_SELF", "b80a9c92d74c5d8067cc7b39e93999ce1c69cd44fa66f46387b863f3a6dc25e0") // not safe!
	viper.SetDefault("ALIENOS_RELAY_PORT", "7771")
	viper.SetDefault("ALIENOS_RELAY_BIND", "0.0.0.0")
	viper.SetDefault("ALIENOS_RELAY_URL", "alienos.jellyfish.land")
	viper.SetDefault("ALIENOS_WHITE_LISTED", false)
	viper.SetDefault("ALIENOS_ADMINS", []string{"badbdda507572b397852048ea74f2ef3ad92b1aac07c3d4e1dec174e8cdc962a"})

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("can't load config: %s", err.Error())
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("can't load config: %s", err.Error())
	}
}
