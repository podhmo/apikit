package classify

import (
	"encoding/json"
	"fmt"
	"io"

	"os"
	"time"
)

type fileinfo struct {
	Name  string    `json:"name"`
	Hash  []byte    `json:"hash"`
	Mtime time.Time `json:"mtime"`
}

type FileEntry struct {
	fileinfo
}

func (e *FileEntry) Name() string {
	return e.fileinfo.Name
}
func (e *FileEntry) String() string {
	return fmt.Sprintf("%s %s", e.fileinfo.Name, e.fileinfo.Hash)
}
func (e *FileEntry) Hash() ([]byte, error) {
	return e.fileinfo.Hash, nil
}
func (e *FileEntry) Mtime() time.Time {
	return e.fileinfo.Mtime
}

type JSONFileStore struct {
	Mtime time.Time
}

func (s *JSONFileStore) WriteFile(filename string, src []Result) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filename, err)
	}
	defer f.Close()
	if err := s.WriteData(f, src); err != nil {
		return fmt.Errorf("encode json %s: %w", filename, err)
	}
	return nil
}

func (s *JSONFileStore) WriteData(w io.Writer, src []Result) error {
	dst := make([]fileinfo, 0, len(src))
	for _, r := range src {
		if r.Type == ResultTypeDelete {
			continue
		}

		hash, err := r.Entry.Hash()
		if err != nil {
			return err
		}

		mtime := s.Mtime
		if r.Type == ResultTypeNotChanged {
			if t, ok := r.Entry.(interface{ Mtime() time.Time }); ok {
				mtime = t.Mtime()
			}
		}
		dst = append(dst, fileinfo{
			Name:  r.Name(),
			Hash:  hash,
			Mtime: mtime,
		})
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(dst); err != nil {
		return fmt.Errorf("encode json %w", err)
	}
	return nil
}

func (s *JSONFileStore) ReadFile(filename string) ([]Entry, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("open %s: %w", filename, err)
	}
	defer f.Close()

	entries, err := s.ReadData(f)
	if err != nil {
		return nil, fmt.Errorf("read data %s: %w", filename, err)
	}
	return entries, nil
}

func (s *JSONFileStore) ReadData(r io.Reader) ([]Entry, error) {
	decoder := json.NewDecoder(r)
	var entries []*FileEntry
	if err := decoder.Decode(&entries); err != nil {
		return nil, err
	}

	dst := make([]Entry, len(entries))
	for i, x := range entries {
		dst[i] = x
	}
	return dst, nil
}
