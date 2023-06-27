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
	Timeout time.Duration // timeout for longpoll
	Path    string        // path to watch
	Logger  Logger        // logger
}

type LongPollImpl struct {
	Config   Config
	quit     chan struct{}
	logger   Logger
	ticker   *time.Ticker
	state    *State
	pathType PathType // path type (file or directory)
}

func NewLongPoll(config Config) LongPoll {
	if config.Logger == nil {
		config.Logger = NewLogger()
	}
	var pathType PathType
	file, err := NewFile(config.Path)
	if err != nil {
		panic(err)
	}

	if file.FileInfo.IsDir() {
		config.Logger.Debugf("path is directory")
		pathType = PathType_DIRECTORY
	} else {
		config.Logger.Debugf("path is file")
		pathType = PathType_FILE
	}

	if config.Timeout == 0 {
		config.Timeout = 1 * time.Second
	}

	return &LongPollImpl{
		Config:   config,
		logger:   config.Logger,
		quit:     make(chan struct{}),
		pathType: pathType,
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
	state, err := l.getContents()
	if err != nil {
		return err
	}

	l.state = state
	return nil
}

func (l *LongPollImpl) eventsFromState(state *State) ([]Event, error) {
	if l.state == nil {
		var events []Event
		for _, file := range state.Files {
			events = append(events, Event{
				Path:      file.AbsolutPath,
				EventType: EventType_CREATE,
			})
		}

		return events, nil
	}

	events := make([]Event, 0)

	for _, file := range state.Files {
		foundFile := l.state.GetFile(file.AbsolutPath)
		if foundFile == nil {
			events = append(events, Event{
				Path:      file.AbsolutPath,
				EventType: EventType_CREATE,
			})
			continue
		}

		if foundFile.FileInfo.ModTime() != file.FileInfo.ModTime() {
			events = append(events, Event{
				Path:      file.AbsolutPath,
				EventType: EventType_MODIFY,
			})
		}

	}

	for _, file := range l.state.Files {
		foundFile := state.GetFile(file.AbsolutPath)
		if foundFile == nil {
			events = append(events, Event{
				Path:      file.AbsolutPath,
				EventType: EventType_DELETE,
			})
		}
	}

	l.state = state
	return events, nil
}

func (l *LongPollImpl) getContents() (*State, error) {
	if l.pathType == PathType_DIRECTORY {
		filePaths, err := l.getContentsDirectory()
		l.logger.Debugf("files %v", filePaths)
		if err != nil {
			return nil, err
		}

		state := NewState()

		for _, filePath := range filePaths {
			file, _ := NewFile(filePath)

			if file == nil {
				continue
			}

			state.AddFile(*file)

			l.logger.Debugf("file %s", file.Name)

		}

		return state, nil

	} else if l.pathType == PathType_FILE {
		file, err := NewFile(l.Config.Path)
		if err != nil {
			return nil, err
		}

		state := NewState()
		state.AddFile(*file)

		return state, nil
	} else {
		// do stuff
	}

	return nil, nil
}

func (l *LongPollImpl) Watch() (<-chan Event, error) {

	err := l.initialScrape()
	if err != nil {
		return nil, err
	}
	eventChan := make(chan Event)

	l.ticker = time.NewTicker(l.Config.Timeout)

	go func() {
		for {
			select {
			case <-l.ticker.C:
				// do stuff
				l.logger.Debugf("longpoll watcher ticked")
				state, err := l.getContents()
				if err != nil {
					l.logger.Errorf("error getting contents: %s", err)
					continue
				}

				events, err := l.eventsFromState(state)

				l.logger.Debugf("events %v", events)

				if err != nil {
					l.logger.Errorf("error processing state: %s", err)
					continue
				}

				for _, event := range events {
					eventChan <- event
				}

			case <-l.quit:
				l.logger.Debugf("longpoll watcher stopping")
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
