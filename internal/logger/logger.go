package logger

import (
	"os"

	"github.com/op/go-logging"
)

func GetLogs(appName string) (log *logging.Logger) {
	log = logging.MustGetLogger("shkaff")
	var format = logging.MustStringFormatter(
		`%{color} %{shortfunc:-9s} %{level:-5s} %{time:15:04:05} %{color:reset} %{message}`,
	)
	appName = ""
	backend := logging.NewLogBackend(os.Stdout, appName, 0)
	backend2Formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backend2Formatter)
	return
}
