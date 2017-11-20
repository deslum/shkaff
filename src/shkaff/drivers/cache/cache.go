package cache

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	CACHEPATH = "./cache/cache.db"
)

type Cache struct {
	*leveldb.DB
}

func InitCacheDB() (cache *Cache) {
	err := os.MkdirAll(CACHEPATH, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	cache = new(Cache)
	cache.DB, err = leveldb.OpenFile(CACHEPATH, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.DB.Close()
	return

}

func (cache *Cache) SetKV(userID, campID int, table, sheet string) (err error) {
	key := []byte(fmt.Sprintf("%d|%d", userID, campID))
	value := []byte(fmt.Sprintf("%s|%s", table, sheet))
	err = cache.DB.Put(key, value, nil)
	if err != nil {
		return err
	}
	return
}

func (cache *Cache) GetKV(userID, campID int) (table, sheet string, err error) {
	var value []byte
	key := []byte(fmt.Sprintf("%d|%d", userID, campID))
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

func (cache *Cache) DeleteKV(userID, campID int) (err error) {
	key := []byte(fmt.Sprintf("%d|%d", userID, campID))
	err = cache.DB.Delete(key, nil)
	if err != nil {
		return err
	}
	return
}

func (cache *Cache) ExistKV(userID, campID int) (result bool, err error) {
	key := []byte(fmt.Sprintf("%d|%d", userID, campID))
	if res, err := cache.Get(key, nil); res == nil && err != nil {
		return false, err
	}
	return true, nil
}
