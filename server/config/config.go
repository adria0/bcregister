package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	DataFolder string
	ServerSecret string
	Recaptcha struct {
		Code string
		Key string
	}
	WebServer struct {
		Prefix string
		WwwRoot string
		Bind string
		CertFile string
		KeyFile string
	}
	SmtpClient struct {
		From string
		Server string
		User string
		Password string
		Domain string
	}
}

var C Config

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("bcserver")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("BCSERVER") 
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		log.Fatal(err)
	}
}