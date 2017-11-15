package databases

import "shkaff/structs"

type DatabaseDriver interface {
	Dump(task *structs.Task)
}
