package source

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/libgox/core/config/dynamic/types"
)

// FileSource implements the Source interface for file-based configuration
type FileSource struct {
	path       string
	sourceType types.SourceType
	watcher    *fsnotify.Watcher
	callback   func([]byte)
	stopCh     chan struct{}
}

func NewFileSource(path string, sourceType types.SourceType) (*FileSource, error) {
	fs := &FileSource{
		path:       path,
		sourceType: sourceType,
		stopCh:     make(chan struct{}),
	}

	return fs, nil
}

func (fs *FileSource) Start() error {
	if fs.sourceType == types.Dynamic {
		var err error
		fs.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		configDir, _ := filepath.Split(filepath.Clean(fs.path))

		err = fs.watcher.Add(configDir)
		if err != nil {
			log.Println("watch file err: ", err)
		}
		go fs.watchUpdates()
	}

	return nil
}

func (fs *FileSource) Stop() error {
	close(fs.stopCh)
	if fs.sourceType == types.Dynamic {
		return fs.watcher.Close()
	}
	return nil
}

func (fs *FileSource) Type() types.SourceType {
	return fs.sourceType
}

func (fs *FileSource) Read() ([]byte, error) {
	return os.ReadFile(fs.path)
}

// SetUpdateCallback registers a callback for dynamic config updates
func (fs *FileSource) SetUpdateCallback(callback func([]byte)) {
	if fs.sourceType != types.Dynamic {
		return
	}
	fs.callback = callback
}

// watchUpdates monitors file changes and triggers callbacks
func (fs *FileSource) watchUpdates() {
	cleanPath := filepath.Clean(fs.path)
	realPath, _ := filepath.EvalSymlinks(fs.path)
	const writeOrCreateMask = fsnotify.Write | fsnotify.Create

	for {
		select {
		case event, ok := <-fs.watcher.Events:
			if !ok {
				return
			}

			if filepath.Clean(event.Name) != cleanPath {
				break
			}

			curRealPath, _ := filepath.EvalSymlinks(fs.path)

			if event.Op&writeOrCreateMask != 0 || (curRealPath != "" && curRealPath != realPath) {
				realPath = curRealPath
				time.Sleep(100 * time.Millisecond)
				configBytes, err := fs.Read()
				if err != nil {
					log.Printf("Failed to read file: %v", err)
					continue
				}

				if fs.callback != nil {
					fs.callback(configBytes)
				}
			} else if event.Op&fsnotify.Remove != 0 {
				if fs.callback != nil {
					fs.callback([]byte(``))
				}
				return
			}

		case err, ok := <-fs.watcher.Errors:
			if ok {
				log.Printf("watcher error: %v\n", err)
			}
			return
		}
	}
}
