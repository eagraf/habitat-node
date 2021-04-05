package state

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
)

// The log is a sequence of serialized base64 encoded JSON objects, with each line being one object
// There is no top level JSON object wrapping it, making it easy to append without scanning all of the data

type LogCollection struct {
	CommunityLogs map[entities.CommunityID]*Log
	NodeLog       *Log
}

type Log struct {
	LogWriter         io.Writer
	CurSequenceNumber int64
}

type Entry struct {
	Transition *transitions.TransitionWrapper

	SequenceNumber int64     `json:"sequence_number"`
	Committed      time.Time `json:"committed"`
}

func NewLog(path string) (*Log, error) {
	walWriter, err := NewWALWriter(path)
	if err != nil {
		return nil, err
	}

	return &Log{
		LogWriter:         walWriter,
		CurSequenceNumber: 0,
	}, nil
}

// WriteAhead appends to the log file. This method should be called before anything else is done to process state
func (l *Log) WriteAhead(transition *transitions.TransitionWrapper) error {
	// Wrap transition in log entry
	entry := &Entry{
		Transition:     transition,
		SequenceNumber: l.CurSequenceNumber,
		Committed:      time.Now(),
	}

	// Marshal JSON
	buf, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Base64 encode
	encoding := base64.StdEncoding.EncodeToString(buf)

	// Append to log file
	_, err = l.LogWriter.Write([]byte(encoding + "\n"))
	if err != nil {
		return err
	}
	l.CurSequenceNumber += 1

	return nil
}

// Helper functions for dealing with the WAL

func DecodeLogEntry(entry []byte) (*Entry, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(entry))
	if err != nil {
		return nil, err
	}

	unmarshalled := &Entry{}
	err = json.Unmarshal(decoded, unmarshalled)
	if err != nil {
		return nil, err
	}

	return unmarshalled, nil
}

type WALWriter struct {
	logPath string
	logFile *os.File
	mutex   *sync.Mutex
}

func (ww *WALWriter) Write(buf []byte) (int, error) {
	ww.mutex.Lock()

	n, err := ww.logFile.Write(buf)
	if err != nil {
		return n, err
	}

	// Call fsync syscall to ensure log is immediately persisted to physical storage
	err = ww.logFile.Sync()
	if err != nil {
		return n, err
	}

	ww.mutex.Unlock()

	return n, nil
}

func NewWALWriter(path string) (*WALWriter, error) {
	// The WAL is kept as a persistently open append and write only file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WALWriter{
		logPath: path,
		logFile: file,
		mutex:   &sync.Mutex{},
	}, nil
}
