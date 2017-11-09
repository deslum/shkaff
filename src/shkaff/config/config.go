package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const (
	CONFIG_FILE                   = "config/shkaff.json"
	DEFAULT_HOST                  = "localhost"
	DEFAULT_DATABASE_PORT         = 5432
	DEFAULT_AMQP_PORT             = 5672
	DEFAULT_DATABASE_DB           = "postgres"
	DEFAULT_REFRESH_DATABASE_SCAN = 60

	INVALID_DATABASE_HOST     = "Database host in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_PORT     = "Database port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_DATABASE_DB       = "Database name in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_USER     = "Database user name is empty"
	INVALID_DATABASE_PASSWORD = "Database password is empty"

	INVALID_AMQP_HOST     = "AMQP host in config file is empty. Shkaff set '%s'\n"
	INVALID_AMQP_PORT     = "AMPQ port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_AMQP_USER     = "AMQP user name is empty"
	INVALID_AMQP_PASSWORD = "AMQP password is empty"
)

type ShkaffConfig struct {
	RMQ_HOST              string `json:"RMQ_HOST"`
	RMQ_PORT              int    `json:"RMQ_PORT"`
	RMQ_USER              string `json:"RMQ_USER"`
	RMQ_PASS              string `json:"RMQ_PASS"`
	RMQ_VHOST             string `json:"RMQ_VHOST"`
	DATABASE_HOST         string `json:"DATABASE_HOST"`
	DATABASE_PORT         int    `json:"DATABASE_PORT"`
	DATABASE_USER         string `json:"DATABASE_USER"`
	DATABASE_PASS         string `json:"DATABASE_PASS"`
	DATABASE_DB           string `json:"DATABASE_DB"`
	DATABASE_SSL          bool   `json:"DATABASE_SSL"`
	REFRESH_DATABASE_SCAN int    `json:"REFRESH_DATABASE_SCAN"`
}

func InitControlConfig() (cc ShkaffConfig) {
	var file []byte
	var err error
	if file, err = ioutil.ReadFile(CONFIG_FILE); err != nil {
		log.Fatalln(err)
		return
	}
	if err := json.Unmarshal(file, &cc); err != nil {
		log.Fatalln(err)
		return
	}
	cc.validate()
	return
}

func (cc *ShkaffConfig) validate() {
	if cc.DATABASE_HOST == "" {
		log.Printf(INVALID_DATABASE_HOST, DEFAULT_HOST)
		cc.DATABASE_HOST = DEFAULT_HOST
	}
	if cc.DATABASE_PORT < 1025 || cc.DATABASE_PORT > 65535 {
		log.Printf(INVALID_DATABASE_PORT, cc.DATABASE_PORT, DEFAULT_DATABASE_PORT)
		cc.DATABASE_PORT = DEFAULT_DATABASE_PORT
	}
	if cc.DATABASE_DB == "" {
		log.Printf(INVALID_DATABASE_DB, DEFAULT_DATABASE_DB)
		cc.DATABASE_DB = DEFAULT_DATABASE_DB
	}
	if cc.DATABASE_USER == "" {
		log.Fatalln(INVALID_DATABASE_USER)
	}
	if cc.DATABASE_PASS == "" {
		log.Fatalln(INVALID_DATABASE_PASSWORD)
	}

	if cc.RMQ_HOST == "" {
		log.Printf(INVALID_AMQP_HOST, DEFAULT_HOST)
		cc.RMQ_HOST = DEFAULT_HOST
	}
	if cc.RMQ_PORT < 1025 || cc.RMQ_PORT > 65535 {
		log.Printf(INVALID_AMQP_PORT, cc.RMQ_PORT, DEFAULT_AMQP_PORT)
		cc.RMQ_PORT = DEFAULT_AMQP_PORT
	}
	if cc.RMQ_USER == "" {
		log.Fatalln(INVALID_AMQP_USER)
	}
	if cc.RMQ_PASS == "" {
		log.Fatalln(INVALID_AMQP_PASSWORD)
	}
	if cc.REFRESH_DATABASE_SCAN == 0 {
		cc.REFRESH_DATABASE_SCAN = DEFAULT_REFRESH_DATABASE_SCAN
	}
	return
}
