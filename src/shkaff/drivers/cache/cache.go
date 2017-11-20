package cache

import(
	"github.com/syndtr/goleveldb/leveldb"
)

type Cache struct {}

func InitCache() (cache *Cache){
	db, err := leveldb.OpenFile("/tmp/foo.db", nil)
	if err != nil {
		log.Fatal("Yikes!")
	}
	defer db.Close()
}