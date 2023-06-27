package filewatcher

import "github.com/fsnotify/fsnotify"

type WatcherType string

const (
	WatcherType_DIRECTORY WatcherType = "DIRECTORY"
	WatcherType_FILE      WatcherType = "FILE"
)

type Watcher interface {
	Close() error
	Watch() error
}

type WatcherImpl struct {
	fsnotifer   *fsnotify.Watcher
	watcherType WatcherType
	path        string
}

func NewWatcher(watcherType WatcherType, path string) Watcher {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		panic(err)
	}

	err = watcher.Add(path)

	if err != nil {
		panic(err)
	}

	return &WatcherImpl{
		fsnotifer:   watcher,
		path:        path,
		watcherType: watcherType,
	}
}

func (w *WatcherImpl) Watch() error {
	return nil
}

func (w *WatcherImpl) Close() error {
	return w.fsnotifer.Close()
}
