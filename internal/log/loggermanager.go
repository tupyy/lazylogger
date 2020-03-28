package log

import (
	"errors"
	"io"

	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/conf"
	"github.com/tupyy/lazylogger/internal/ssh"
)

// maxiumum amount of data requested when a writer is registered
const RequestDataMaxSize = 1024 * 150 // 150 kB

// LoggerManager handles the loggers and write any new data received from loggers to Gui LogWriter implementations.
// A logger is created (and ssh connection dialed) only when is registered to a view. It will still run after the view unregistered.
type LoggerManager struct {

	// map of loggers
	loggers map[int]*Logger

	// manage ssh connections
	sshPool *ssh.SSHPool

	// channel to received data notification from loggers
	in chan interface{}

	// holds a map with registered writers. A writer can represent a textview or stdout
	writers map[int][]LogWriter

	done chan interface{}

	configurations map[int]conf.LoggerConfiguration
}

// LogWriter is an interface that extends Writer interface by adding
// a method to handle the change in state of the logger.
type LogWriter interface {
	io.Writer
	SetState(state string, err error)
}

func NewLoggerManager(configurations []conf.LoggerConfiguration) *LoggerManager {

	lm := &LoggerManager{
		loggers:        make(map[int]*Logger),
		sshPool:        ssh.NewSSHPool(),
		in:             make(chan interface{}),
		writers:        make(map[int][]LogWriter),
		done:           make(chan interface{}),
		configurations: mapFromArray(configurations),
	}

	return lm
}

// CreateLoggers returns the number of loggers created.
func (lm *LoggerManager) CreateLogger(id int, conf conf.LoggerConfiguration) (*Logger, error) {

	client, err := lm.sshPool.Connect(conf)
	if err != nil {
		return nil, err
	}

	remoteReader := NewRemoteReader(client, conf.File)
	logger := NewLogger(id, lm.in)
	logger.Start(remoteReader)
	lm.loggers[id] = logger

	return logger, nil
}

func (lm *LoggerManager) GetConfigurations() map[int]conf.LoggerConfiguration {
	return lm.configurations
}

func (lm *LoggerManager) Run() {
	for {
		select {
		case n := <-lm.in:
			switch v := n.(type) {
			case DataNotification:
				glog.V(3).Infof("DataNotification received from %d", v.ID)
				data, _ := lm.RequestData(v.ID, v.PreviousSize, int(v.Size-v.PreviousSize))
				if w, ok := lm.writers[v.ID]; ok {
					for _, l := range w {
						l.Write(data)
					}
				}
			case State:
				if w, ok := lm.writers[v.ID]; ok {
					for _, l := range w {
						l.SetState(v.String(), v.Err)
					}
				}
			}
		case <-lm.done:
			return
		}
	}
}

func (lm *LoggerManager) RegisterWriter(loggerID int, w LogWriter) error {
	logger, err := lm.getLogger(loggerID)
	if err != nil {
		return err
	}

	if ww, ok := lm.writers[loggerID]; ok {
		ww = append(ww, w)
		lm.writers[loggerID] = ww
	} else {
		lm.writers[loggerID] = []LogWriter{w}
	}

	requestSize := logger.CacheSize()
	offset := 0
	if logger.CacheSize() > RequestDataMaxSize {
		requestSize = RequestDataMaxSize
		offset = logger.CacheSize() - RequestDataMaxSize
	}

	data, _ := logger.RequestData(int64(offset), requestSize)
	w.Write(data)

	w.SetState(logger.State.String(), logger.State.Err)
	return nil
}

func (lm *LoggerManager) UnregisterWriter(loggerID int) error {

	if _, ok := lm.writers[loggerID]; ok {
		delete(lm.writers, loggerID)
	}

	return nil
}

func (lm *LoggerManager) RequestData(id int, offset int64, size int) ([]byte, error) {
	logger, ok := lm.loggers[id]
	if !ok {
		return []byte{}, errors.New("logger not found")
	}

	data, _ := logger.RequestData(offset, size)
	return data, nil
}

// Close all the loggers
func (lm *LoggerManager) Stop() {
	for _, logger := range lm.loggers {
		logger.Stop()
		logger = nil
	}

	lm.loggers = make(map[int]*Logger)
	lm.done <- struct{}{}
}

// StopLogger stops a loggers and returns its id if service found.
func (lm *LoggerManager) stopLogger(id int) int {
	logger, ok := lm.loggers[id]
	if ok {
		glog.Infof("Stopping logger %d", logger.ID)
		logger.Stop()
		delete(lm.loggers, id)
		return logger.ID
	}

	return id
}

// Return or create a logger
func (lm *LoggerManager) getLogger(id int) (*Logger, error) {
	l, ok := lm.loggers[id]
	if !ok {
		conf := lm.configurations[id]
		logger, err := lm.CreateLogger(id, conf)
		if err != nil {
			return nil, err
		}
		lm.loggers[id] = logger
		l = logger
	}

	return l, nil
}
