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

	cfg := zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding: "json",
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	logger, _ := cfg.Build()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	longpollWatcher := longpoll.NewLongPoll(longpoll.Config{
		Timeout: 1 * time.Second,
		Path:    "./tmp",
		Logger:  sugar,
	})

	watcherChan, _ := longpollWatcher.Watch()

	go func() {
		for {
			select {
			case event := <-watcherChan:
				if event.EventType == longpoll.EventType_CREATE {
					sugar.Infof("create file: %v", event)
				}

				if event.EventType == longpoll.EventType_DELETE {
					sugar.Infof("delete file: %v", event)
				}

				if event.EventType == longpoll.EventType_MODIFY {
					sugar.Infof("modify file: %v", event)
				}

			}
		}

	}()

	time.Sleep(20 * time.Second)

	longpollWatcher.Close()
	time.Sleep(2 * time.Second)
}
