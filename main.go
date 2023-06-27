package main

import (
	"configReloader/longpoll"
	"go.uber.org/zap"
	"time"
)

func main() {

	//watcher := NewWatcher(WatcherType_DIRECTORY, "/tmp")
	//
	//defer watcher.Close()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	longpollWatcher := longpoll.NewLongPoll(longpoll.Config{
		Timeout:  1 * time.Second,
		Path:     "./tmp",
		PathType: longpoll.PathType_DIRECTORY,
		Logger:   sugar,
	})

	watcherChan, _ := longpollWatcher.Watch()

	go func() {
		for {
			select {
			case event := <-watcherChan:
				sugar.Infof("event: %v", event)

			case <-time.After(5 * time.Second):
				sugar.Infof("timeout")
			}
		}

	}()

	time.Sleep(10 * time.Second)

	longpollWatcher.Close()
	time.Sleep(2 * time.Second)
}
