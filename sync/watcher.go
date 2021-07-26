package sync

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	basePath string
	// Tracks whether this is the first pass of the directory tree. If not,
	// then any new file encountered will be treated as an update.
	firstPass  bool
	fileHashes map[string]string
}

func (w *Watcher) getFileHash(path string) (string, error) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	hasher.Write([]byte(fileData))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (w *Watcher) postFileIfChanged(path string) error {
	_, file := filepath.Split(path)
	fileSplit := strings.Split(file, ".")

	// Ignore files with no extension
	if len(fileSplit) < 2 {
		return nil
	}

	if !contains(getWatchedExts(), fileSplit[1]) {
		return nil
	}

	hashed, err := w.getFileHash(path)
	if err != nil {
		return err
	}

	if !w.firstPass {
		oldHash := w.fileHashes[path]
		if oldHash != hashed {
			err := SendFileUpdate(path)

			if err != nil {
				return err
			}
		}
	}
	w.fileHashes[path] = hashed
	return nil
}

func (w *Watcher) checkDirForChanges(dirPath string) error {
	dir, _ := filepath.Split(dirPath)

	if contains(getIgnoredDirs(), dir) {
		return nil
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		newPath := path.Join(dirPath, fileInfo.Name())
		if fileInfo.IsDir() {
			// Recursively check subdirectories
			err = w.checkDirForChanges(newPath)
		} else {
			err = w.postFileIfChanged(newPath)
		}
	}
	return err
}

func (w *Watcher) PollForChanges(waitTime time.Duration) {
	for {
		err := w.checkDirForChanges(w.basePath)
		if err != nil {
			// This should not be reached except in exceptional circumstances
			// We want to continue looping even if we hit an unexpected error
			log.Println("Unexpected error in file watcher:", err)
		} else {
			w.firstPass = false
		}
		// Poll for changes every `waitTime` seconds, whether or not the
		// previous call succeeded.
		time.Sleep(waitTime)
	}
}

func NewWatcher(basePath string) *Watcher {
	return &Watcher{
		basePath:   basePath,
		firstPass:  true,
		fileHashes: map[string]string{},
	}
}

func contains(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}
