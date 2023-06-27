package longpoll

import (
	"os"
	"time"
)

type PathType string
type EventType string

type Logger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
}

const (
	PathType_DIRECTORY PathType = "DIRECTORY"
	PathType_FILE      PathType = "FILE"

	EventType_CREATE EventType = "CREATE"
	EventType_DELETE EventType = "DELETE"
	EventType_MODIFY EventType = "MODIFY"
)

type Event struct {
	Path      string
	EventType EventType
}

type LongPoll interface {
	Close() error
	// start watcher return chanell of events
	Watch() (<-chan Event, error)
}

type Config struct {
	Timeout  time.Duration // timeout for longpoll
	Path     string        // path to watch
	PathType PathType      // path type (file or directory)
	Logger   Logger        // logger
}

type LongPollImpl struct {
	Config Config
	quit   chan struct{}
	logger Logger
	ticker *time.Ticker
}

func NewLongPoll(config Config) LongPoll {
	return &LongPollImpl{
		Config: config,
		logger: config.Logger,
		quit:   make(chan struct{}),
	}
}

func (l *LongPollImpl) getContentsDirectory() ([]string, error) {
	files, err := os.ReadDir(l.Config.Path)

	if err != nil {
		return nil, err
	}

	var fileNames []string

	for _, file := range files {
		fileNames = append(fileNames, l.Config.Path+"/"+file.Name())
	}

	return fileNames, nil
}

func (l *LongPollImpl) initialScrape() error {

}

func (l *LongPollImpl) getContents() ([]Event, error) {
	if l.Config.PathType == PathType_DIRECTORY {
		files, err := l.getContentsDirectory()
		l.logger.Infof("files %v", files)
		if err != nil {
			return nil, err
		}

		var events []Event

		for _, file := range files {
			fileStats, _ := os.Stat(file)
			if fileStats == nil {
				continue
			}
			l.logger.Infof("file %s", fileStats.Name())
			events = append(events, Event{
				Path:      file,
				EventType: EventType_CREATE,
			})
		}

		return events, nil

	} else if l.Config.PathType == PathType_FILE {
		// do stuff
	} else {
		// do stuff
	}

	return nil, nil
}

func (l *LongPollImpl) Watch() (<-chan Event, error) {

	eventChan := make(chan Event)

	l.ticker = time.NewTicker(l.Config.Timeout)

	go func() {
		for {
			select {
			case <-l.ticker.C:
				// do stuff
				l.logger.Infof("longpoll watcher ticked")
				events, err := l.getContents()

				if err != nil {
					l.logger.Errorf("error getting contents: %s", err)
					continue
				}

				for _, event := range events {
					eventChan <- event
				}

			case <-l.quit:
				l.logger.Infof("longpoll watcher stopping")
				l.ticker.Stop()
				return
			}
		}
	}()

	return eventChan, nil
}

func (l *LongPollImpl) Close() error {
	l.quit <- struct{}{}
	return nil
}
