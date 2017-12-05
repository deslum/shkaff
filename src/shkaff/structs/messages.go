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
	TaskID      int       `json:"task_id" db:"task_id"`
	DBID        int       `json:"db_id" db:"db_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Databases   string    `json:"-" db:"databases"`
	DBType      string    `json:"-" db:"db_type"`
	Verb        int       `json:"verb" db:"verb"`
	ThreadCount int       `json:"thread_count" db:"thread_count"`
	Gzip        bool      `json:"gzip" db:"gzip"`
	Ipv6        bool      `json:"ipv6" db:"ipv6"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	StartTime   time.Time `json:"start_time" db:"start_time"`
	DBUser      string    `json:"db_user" db:"db_user"`
	DBPassword  string    `json:"db_password" db:"db_password"`
	Database    string    `json:"database"`
	Sheet       string    `json:"sheet"`
}

type StatMessage struct {
	TaskID int    `json:"task_id"`
	DBID   int    `json:"db_id"`
	UserID int    `json:"user_id"`
	Act    Action `json:"action"`
	Error  error  `json:"error"`
}
