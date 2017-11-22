package logger

import (
	"os"

	"github.com/op/go-logging"
)

func InitLogger(prefix string) (log *logging.Logger) {
	log = logging.MustGetLogger("example")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend)
	logging.AddModuleLevel(backend1Leveled)
	return
}
