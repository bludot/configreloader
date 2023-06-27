package longpoll_test

import (
	"configReloader/longpoll"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type eventsMutex struct {
	sync.Mutex
	create int
	delete int
	modify int
}

func (e *eventsMutex) addCreate() {
	e.Lock()
	defer e.Unlock()
	e.create++
}

func (e *eventsMutex) addDelete() {
	e.Lock()
	defer e.Unlock()
	e.delete++
}

func (e *eventsMutex) addModify() {
	e.Lock()
	defer e.Unlock()
	e.modify++
}

func TestNewLongPoll(t *testing.T) {
	f, _ := os.Getwd()
	// root of project
	basepath := filepath.Dir(f)
	t.Run("test", func(t *testing.T) {
		a := assert.New(t)

		longpollWatcher := longpoll.NewLongPoll(longpoll.Config{
			Timeout: 1 * time.Second,
			Path:    basepath + "/test",
		})

		a.NotNil(longpollWatcher)
	})
}

func TestLongPollImpl_Watch(t *testing.T) {

	f, _ := os.Getwd()
	// root of project
	basepath := filepath.Dir(f)
	t.Run("test folder files", func(t *testing.T) {
		a := assert.New(t)

		// create a file
		_, _ = os.Create(basepath + "/test/test.txt")

		longpollWatcher := longpoll.NewLongPoll(longpoll.Config{
			Timeout: 100 * time.Millisecond,
			Path:    basepath + "/test",
		})

		watcherChan, err := longpollWatcher.Watch()
		a.Nil(err)

		events := &eventsMutex{}

		quit := make(chan bool)
		go func() {
			for {
				select {
				case event := <-watcherChan:
					if event.EventType == longpoll.EventType_CREATE {
						events.addCreate()
					}

					if event.EventType == longpoll.EventType_DELETE {
						events.addDelete()
					}

					if event.EventType == longpoll.EventType_MODIFY {
						events.addModify()
					}
				case <-quit:
					return
				}
			}
		}()

		_, _ = os.Create(basepath + "/test/test2.txt")
		time.Sleep(1 * time.Second)

		// modify a file
		_ = os.Chtimes(basepath+"/test/test.txt", time.Now(), time.Now())
		time.Sleep(1 * time.Second)

		// delete a file
		_ = os.Remove(basepath + "/test/test.txt")
		time.Sleep(1 * time.Second)

		// rename a file
		_ = os.Rename(basepath+"/test/test2.txt", basepath+"/test/test3.txt")
		time.Sleep(1 * time.Second)

		_ = os.Remove(basepath + "/test/test3.txt")
		time.Sleep(1 * time.Second)

		quit <- true

		err = longpollWatcher.Close()

		a.Nil(err)
		time.Sleep(1 * time.Second)

		a.Equal(2, events.create)
		a.Equal(3, events.delete)
		a.Equal(1, events.modify)

	})
}
