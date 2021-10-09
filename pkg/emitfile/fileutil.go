package emitfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

var mkdirSentinelMap = map[string]bool{}

func WriteOrCreateFile(path string, b []byte, config *Config) error {
	defer func() {
		if config.Verbose {
			config.Log.Printf("\tF create %s", path) // todo: detect Create Or Update Or Delete (?)
		}
	}()

	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		dirpath := filepath.Dir(path)
		if _, ok := mkdirSentinelMap[dirpath]; ok {
			return err
		}

		mkdirSentinelMap[dirpath] = true
		config.Log.Printf("\tD create %s", dirpath)
		if err := os.MkdirAll(dirpath, 0744); err != nil {
			config.Log.Printf("ERROR: %s", err)
			return err
		}
		return ioutil.WriteFile(path, b, 0666)
	}
	return nil
}
