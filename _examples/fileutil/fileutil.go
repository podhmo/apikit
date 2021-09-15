package fileutil

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var mkdirSentinelMap = map[string]bool{}

func WriteOrCreateFile(path string, b []byte) error {
	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		dirpath := filepath.Dir(path)
		if _, ok := mkdirSentinelMap[dirpath]; ok {
			return err
		}

		mkdirSentinelMap[dirpath] = true
		log.Printf("INFO: directory is not found, try to create %s", dirpath)
		if err := os.MkdirAll(dirpath, 0744); err != nil {
			log.Printf("ERROR: %s", err)
			return err
		}
		return ioutil.WriteFile(path, b, 0666)
	}
	return nil
}
