package emitfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type fileSaver struct {
	config           *Config
	mkdirSentinelMap map[string]bool
}

func newfileSaver(config *Config) *fileSaver {
	return &fileSaver{
		config:           config,
		mkdirSentinelMap: map[string]bool{},
	}
}

func (wf *fileSaver) SaveOrCreateFile(path string, b []byte, prefix string) error {
	defer func() {
		if wf.config.Verbose {
			relative, err := filepath.Rel(wf.config.CurDir, path)
			if err == nil {
				path = relative
			}
			wf.config.Log.Printf("\t%s file %s", prefix, path) // todo: detect Create Or Update Or Delete (?)
		}
	}()

	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		dirpath := filepath.Dir(path)
		if _, ok := wf.mkdirSentinelMap[dirpath]; ok {
			return err
		}

		wf.mkdirSentinelMap[dirpath] = true
		if wf.config.Verbose {
			wf.config.Log.Printf("\t%s directory %s", prefix, dirpath)
		}
		if err := os.MkdirAll(dirpath, 0744); err != nil {
			wf.config.Log.Printf("ERROR: %s", err)
			return err
		}
		return ioutil.WriteFile(path, b, 0666)
	}
	return nil
}
