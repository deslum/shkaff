package databases

import "shkaff/structs"

type DatabaseDriver interface {
	Dump(task *structs.Task) (err error)
	Restore(task *structs.Task) (err error)
}
