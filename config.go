package main

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	WorkingDirectory string `mapstructure:"ALIENOS_WORK_DIR"`

	RelayName        string `mapstructure:"ALIENOS_RELAY_NAME"`
	RelayIcon        string `mapstructure:"ALIENOS_RELAY_ICON"`
	RelayBanner      string `mapstructure:"ALIENOS_RELAY_BANNER"`
	RelayDescription string `mapstructure:"ALIENOS_RELAY_DESCRIPTION"`
	RelayPublicKey   string `mapstructure:"ALIENOS_RELAY_PUBKEY"`
	RelayContact     string `mapstructure:"ALIENOS_RELAY_CONTACT"`
	RelaySelf        string `mapstructure:"ALIENOS_RELAY_SELF"`
	RelayPort        int    `mapstructure:"ALIENOS_RELAY_PORT"`
	RelayBind        string `mapstructure:"ALIENOS_RELAY_BIND"`
	RelayURL         string `mapstructure:"ALIENOS_RELAY_URL"`

	WhiteListedPubkey bool `mapstructure:"ALIENOS_PUBKEY_WHITE_LISTED"`
	WhiteListedKind   bool `mapstructure:"ALIENOS_KIND_WHITE_LISTED"`

	BackupEnabled  bool   `mapstructure:"ALIENOS_BACKUP_ENABLE"`
	BackupInterval int    `mapstructure:"ALIENOS_BACKUP_INTERVAL_HOURS"`
	S3AccessKeyID  string `mapstructure:"ALIENOS_S3_ACCESS_KEY_ID"`
	S3SecretKey    string `mapstructure:"ALIENOS_S3_SECRET_KEY"`
	S3Endpoint     string `mapstructure:"ALIENOS_S3_ENDPOINT"`
	S3Region       string `mapstructure:"ALIENOS_S3_REGION"`
	S3BucketName   string `mapstructure:"ALIENOS_S3_BUCKET_NAME"`

	S3ForBlossom    bool   `mapstructure:"ALIENOS_S3_AS_BLOSSOM_STORAGE"`
	S3BlossomBucket string `mapstructure:"ALIENOS_S3_BLOSSOM_BUCKET"`

	Admins []string `mapstructure:"ALIENOS_ADMINS"`

	LogFilename     string   `mapstructure:"ALIENOS_LOG_FILENAME"`
	LogLevel        string   `mapstructure:"ALIENOS_LOG_LEVEL"`
	LogTargets      []string `mapstructure:"ALIENOS_LOG_TARGETS"`
	LogFileMaxSize  int      `mapstructure:"ALIENOS_LOG_MAX_SIZE"`
	LogFileCompress bool     `mapstructure:"ALIENOS_LOG_FILE_COMPRESS"`
}

func LoadConfig() {
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")

	viper.SetDefault("ALIENOS_WORK_DIR", "alienos_wd/")
	viper.SetDefault("ALIENOS_RELAY_NAME", "alienos")
	viper.SetDefault("ALIENOS_RELAY_ICON", "https://raw.githubusercontent.com/dezh-tech/alienos/refs/heads/main/.image/logo.png?token=GHSAT0AAAAAACYQ42AVUYV5HPNY2PCL4PR6Z5HXGAA")
	viper.SetDefault("ALIENOS_RELAY_BANNER", "https://raw.githubusercontent.com/dezh-tech/alienos/d11be85135ce5dddcc4350d8a779396761642d76/.image/banner.png?token=GHSAT0AAAAAACYQ42AVV3LOPA6HWE5JKL2MZ5HXFRQ")
	viper.SetDefault("ALIENOS_RELAY_DESCRIPTION", "A self-hosting Nostr stack!")
	viper.SetDefault("ALIENOS_RELAY_PUBKEY", "badbdda507572b397852048ea74f2ef3ad92b1aac07c3d4e1dec174e8cdc962a")
	viper.SetDefault("ALIENOS_RELAY_CONTACT", "hi@dezh.tech")
	viper.SetDefault("ALIENOS_RELAY_SELF", "b80a9c92d74c5d8067cc7b39e93999ce1c69cd44fa66f46387b863f3a6dc25e0") // not safe!
	viper.SetDefault("ALIENOS_RELAY_PORT", 7771)
	viper.SetDefault("ALIENOS_RELAY_BIND", "0.0.0.0")
	viper.SetDefault("ALIENOS_RELAY_URL", "alienos.jellyfish.land")

	viper.SetDefault("ALIENOS_PUBKEY_WHITE_LISTED", false)
	viper.SetDefault("ALIENOS_KIND_WHITE_LISTED", false)

	viper.SetDefault("ALIENOS_ADMINS", []string{"badbdda507572b397852048ea74f2ef3ad92b1aac07c3d4e1dec174e8cdc962a"})

	viper.SetDefault("ALIENOS_BACKUP_ENABLE", false)
	viper.SetDefault("ALIENOS_S3_AS_BLOSSOM_STORAGE", false)

	viper.SetDefault("ALIENOS_LOG_FILENAME", "alienos.log")
	viper.SetDefault("ALIENOS_LOG_LEVEL", "info")
	viper.SetDefault("ALIENOS_LOG_TARGETS", []string{"file", "console"})
	viper.SetDefault("ALIENOS_LOG_MAX_SIZE", 10)
	viper.SetDefault("ALIENOS_LOG_FILE_COMPRESS", true)

	viper.AutomaticEnv()

	if _, err := os.Stat(".env"); err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("can't load config: %s", err.Error())
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("can't load config: %s", err.Error())
	}
}
