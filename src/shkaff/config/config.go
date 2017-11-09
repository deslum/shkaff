package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"shkaff/consts"
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
	if file, err = ioutil.ReadFile(consts.CONFIG_FILE); err != nil {
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
		log.Printf(consts.INVALID_DATABASE_HOST, consts.DEFAULT_HOST)
		cc.DATABASE_HOST = consts.DEFAULT_HOST
	}
	if cc.DATABASE_PORT < 1025 || cc.DATABASE_PORT > 65535 {
		log.Printf(consts.INVALID_DATABASE_PORT, cc.DATABASE_PORT, consts.DEFAULT_DATABASE_PORT)
		cc.DATABASE_PORT = consts.DEFAULT_DATABASE_PORT
	}
	if cc.DATABASE_DB == "" {
		log.Printf(consts.INVALID_DATABASE_DB, consts.DEFAULT_DATABASE_DB)
		cc.DATABASE_DB = consts.DEFAULT_DATABASE_DB
	}
	if cc.DATABASE_USER == "" {
		log.Fatalln(consts.INVALID_DATABASE_USER)
	}
	if cc.DATABASE_PASS == "" {
		log.Fatalln(consts.INVALID_DATABASE_PASSWORD)
	}

	if cc.RMQ_HOST == "" {
		log.Printf(consts.INVALID_AMQP_HOST, consts.DEFAULT_HOST)
		cc.RMQ_HOST = consts.DEFAULT_HOST
	}
	if cc.RMQ_PORT < 1025 || cc.RMQ_PORT > 65535 {
		log.Printf(consts.INVALID_AMQP_PORT, cc.RMQ_PORT, consts.DEFAULT_AMQP_PORT)
		cc.RMQ_PORT = consts.DEFAULT_AMQP_PORT
	}
	if cc.RMQ_USER == "" {
		log.Fatalln(consts.INVALID_AMQP_USER)
	}
	if cc.RMQ_PASS == "" {
		log.Fatalln(consts.INVALID_AMQP_PASSWORD)
	}
	if cc.REFRESH_DATABASE_SCAN == 0 {
		cc.REFRESH_DATABASE_SCAN = consts.DEFAULT_REFRESH_DATABASE_SCAN
	}
	return
}
