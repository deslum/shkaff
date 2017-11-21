package cache

import (
	"errors"
	"fmt"
	"log"
	"os"
	"shkaff/consts"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

type Cache struct {
	*leveldb.DB
}

func InitCacheDB() (cache *Cache) {
	err := os.MkdirAll(consts.CACHEPATH, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	cache = new(Cache)
	cache.DB, err = leveldb.OpenFile(consts.CACHEPATH, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.DB.Close()
	return

}

func (cache *Cache) SetKV(userID, dbID, taskID int, table, sheet string) (err error) {
	key := []byte(fmt.Sprintf("%d|%d|%d", userID, dbID, taskID))
	value := []byte(fmt.Sprintf("%s|%s", table, sheet))
	err = cache.DB.Put(key, value, nil)
	if err != nil {
		return err
	}
	return
}

func (cache *Cache) GetKV(userID, dbID, taskID int) (table, sheet string, err error) {
	var value []byte
	key := []byte(fmt.Sprintf("%d|%d|%d", userID, dbID, taskID))
	value, err = cache.DB.Get(key, nil)
	if err != nil {
		return "", "", err
	}
	valueStr := string(value[:])
	val := strings.Split(valueStr, "|")
	if len(val) != 2 {
		log.Printf("Value %s in key %s is bad\n", valueStr, key)
		cache.Delete(key, nil)
		return "", "", errors.New("Bad key-value")
	}
	table = val[0]
	sheet = val[1]
	return table, sheet, nil
}

func (cache *Cache) DeleteKV(userID, dbID, taskID int) (err error) {
	key := []byte(fmt.Sprintf("%d|%d|%d", userID, dbID, taskID))
	err = cache.DB.Delete(key, nil)
	if err != nil {
		return err
	}
	return
}

func (cache *Cache) ExistKV(userID, dbID, taskID int) (result bool, err error) {
	key := []byte(fmt.Sprintf("%d|%d|%d", userID, dbID, taskID))
	res, err := cache.Get(key, nil)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, err
	}
	return true, nil
}
