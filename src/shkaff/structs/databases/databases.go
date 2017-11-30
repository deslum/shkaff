package databases

import "shkaff/structs"

type DatabaseDriver interface {
	Dump(task *structs.Task) (dumpResult string, err error)
	Restore(task *structs.Task) (dumpResult string, err error)
}
