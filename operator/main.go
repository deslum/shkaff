package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	CONFIG_FILE           = "operator.json"
	DEFAULT_HOST          = "localhost"
	DEFAULT_DATABASE_PORT = 5432
	DEFAULT_AMQP_PORT     = 5672
	DEFAULT_DATABASE_DB   = "postgres"

	INVALID_DATABASE_HOST = "Database host in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_PORT = "Database port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_DATABASE_DB   = "Database name in config file is empty. Shkaff set '%s'\n"

	INVALID_AMQP_HOST = "AMQP host in config file is empty. Shkaff set '%s'\n"
	INVALID_AMQP_PORT = "AMPQ port %d in config file invalid. Shkaff set '%d'\n"

	URI_TEMPLATE = "%s://%s:%s@%s:%d/%s"
	POSTGRES     = "postgres"
	AMQP         = "amqp"
)

type ControlConfig struct {
	RMQ_HOST      string `json:"RMQ_HOST"`
	RMQ_PORT      int    `json:"RMQ_PORT"`
	RMQ_USER      string `json:"RMQ_USER"`
	RMQ_PASS      string `json:"RMQ_PASS"`
	RMQ_VHOST     string `json:"RMQ_VHOST"`
	DATABASE_HOST string `json:"DATABASE_HOST"`
	DATABASE_PORT int    `json:"DATABASE_PORT"`
	DATABASE_USER string `json:"DATABASE_USER"`
	DATABASE_PASS string `json:"DATABASE_PASS"`
	DATABASE_DB   string `json:"DATABASE_DB"`
}

type pSQL struct {
	uri string
}

type amqp struct {
	uri string
}

func initControlConfig(filename string) (cc ControlConfig) {
	var file []byte
	var err error
	if file, err = ioutil.ReadFile(filename); err != nil {
		log.Fatalln(err)
		return
	}
	if err := json.Unmarshal(file, &cc); err != nil {
		log.Fatalln(err)
		return
	}
	return
}

func (cc *ControlConfig) validateConfig() (isValid bool) {
	isValid = true
	if cc.DATABASE_HOST == "" {
		log.Printf(INVALID_DATABASE_HOST, DEFAULT_HOST)
		cc.DATABASE_HOST = DEFAULT_HOST
	}
	if cc.DATABASE_PORT < 1025 && cc.DATABASE_PORT > 65535 {
		log.Printf(INVALID_DATABASE_PORT, cc.DATABASE_PORT, DEFAULT_DATABASE_PORT)
		cc.DATABASE_PORT = DEFAULT_DATABASE_PORT
	}
	if cc.DATABASE_DB == "" {
		log.Printf(INVALID_DATABASE_DB, DEFAULT_DATABASE_DB)
		cc.DATABASE_DB = DEFAULT_DATABASE_DB
	}
	if cc.RMQ_HOST == "" {
		log.Printf(INVALID_AMQP_HOST, DEFAULT_HOST)
		cc.RMQ_HOST = DEFAULT_HOST
	}
	if cc.RMQ_PORT < 1025 && cc.RMQ_PORT > 65535 {
		log.Printf(INVALID_AMQP_PORT, cc.RMQ_PORT, DEFAULT_AMQP_PORT)
		cc.RMQ_PORT = DEFAULT_AMQP_PORT
	}
	return
}

func initPSQL(cf ControlConfig) (ps *pSQL) {
	ps = new(pSQL)
	ps.uri = fmt.Sprintf(URI_TEMPLATE, POSTGRES,
		cf.DATABASE_USER,
		cf.DATABASE_PASS,
		cf.DATABASE_HOST,
		cf.DATABASE_PORT,
		cf.DATABASE_DB)
	return
}

func initAMQP(cf ControlConfig) (qp *amqp) {
	qp = new(amqp)
	qp.uri = fmt.Sprintf(URI_TEMPLATE, AMQP,
		cf.RMQ_USER,
		cf.RMQ_PASS,
		cf.RMQ_HOST,
		cf.RMQ_PORT,
		cf.RMQ_VHOST)
	return
}

func main() {
	controlConfig := initControlConfig(CONFIG_FILE)
	if controlConfig.validateConfig() {
		pSQL := initPSQL(controlConfig)
		amqp := initAMQP(controlConfig)
		fmt.Println(pSQL.uri)
		fmt.Println(amqp.uri)
	}
}
