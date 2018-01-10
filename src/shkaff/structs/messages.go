package structs

import (
	"time"
)

type Action int

// 0 - StartDumping
// 1 - SuccessDumping
// 2 - FailDumping
// 3 - StartRestoring
// 4 - SuccessRestoring
// 5 - FailRestoring

const (
	NewOperator Action = 0 + iota
	StartOperator
	SuccessOperator
	FailOperator
	NewDump
	StartDump
	SuccessDump
	FailDump
	NewRestore
	StartRestore
	SuccessRestore
	FailRestore
)

type Task struct {
	DBSettingsType int    `json:"-" db:"type_id"`
	DBSettingsID   int    `json:"-" db:"db_settings_id"`
	IsActive       bool   `json:"-" db:"is_active"`
	TaskID         int    `json:"task_id" db:"task_id"`
	TaskName       string `json:"task_name" db:"task_name"`
	DBID           int    `json:"db_id" db:"db_id"`
	UserID         int    `json:"user_id" db:"user_id"`
	Databases      string `json:"-" db:"databases"`
	DBType         string `json:"-" db:"db_type"`
	Verb           int    `json:"verb" db:"verbose"`
	ThreadCount    int    `json:"thread_count" db:"thread_count"`
	Gzip           bool   `json:"gzip" db:"gzip"`
	Ipv6           bool   `json:"ipv6" db:"ipv6"`
	Host           string `json:"host" db:"host"`
	Port           int    `json:"port" db:"port"`
	DBUser         string `json:"db_user" db:"db_user"`
	DBPassword     string `json:"db_password" db:"db_password"`
	Database       string `json:"database"`
	Sheet          string `json:"sheet"`
}

type APITask struct {
	TaskID      int         `db:"task_id"`
	TaskName    string      `db:"task_name"`
	IsActive    bool        `db:"is_active"`
	DBID        int         `db:"db_id"`
	Databases   string      `db:"databases"`
	Verb        int         `db:"verbose"`
	ThreadCount int         `db:"thread_count"`
	Gzip        bool        `db:"gzip"`
	Ipv6        bool        `db:"ipv6"`
	Months      string      `db:"months"`
	Days        string `db:"days"`
	Hours       string `db:"hours"`
	Minutes     string `db:"minutes"`
}

type StatMessage struct {
	UserID          uint16    `db:"UserId" json:"uid"`
	DbID            uint16    `db:"DbID" json:"did"`
	TaskID          uint16    `db:"TaskId" json:"tid"`
	NewOperator     uint32    `db:"NewOperator" json:"no"`
	SuccessOperator uint32    `db:"SuccessOperator" json:"so"`
	FailOperator    uint32    `db:"FailOperator" json:"fo"`
	ErrorOperator   string    `db:"ErrorOperator" json:"eo"`
	NewDump         uint32    `db:"NewDump" json:"nd"`
	SuccessDump     uint32    `db:"SuccessDump" json:"sd"`
	FailDump        uint32    `db:"FailDump" json:"fd"`
	ErrorDump       string    `db:"ErrorDump" json:"ed"`
	NewRestore      uint32    `db:"NewRestore" json:"nr"`
	SuccessRestore  uint32    `db:"SuccessRestore" json:"sr"`
	FailRestore     uint32    `db:"FailRestore" json:"fr"`
	ErrorRestore    string    `db:"ErrorRestore" json:"er"`
	CreateDate      time.Time `db:"CreateDate" json:"cd"`
}
