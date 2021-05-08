package state

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
	CurSequenceNumber uint64
	Path              string

	logReader io.Reader
	logWriter io.Writer
	mutex     *sync.Mutex
}

type Entry struct {
	Transition *transitions.TransitionWrapper

	SequenceNumber uint64    `json:"sequence_number"`
	Committed      time.Time `json:"committed"`
}

func NewLog(path string) (*Log, error) {
	walWriter, err := NewWALWriter(path)
	if err != nil {
		return nil, err
	}

	return &Log{
		CurSequenceNumber: 0,
		Path:              path,

		logWriter: walWriter,
		mutex:     &sync.Mutex{},
	}, nil
}

// WriteAhead appends to the log file. This method should be called before anything else is done to process state
// TODO AHJEWIOGJEIOJGIOEWJGIOEJNDWGIOJNEIO sequence number should be passed and should match cur sequence number ajewoigjnweoigf
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

	logLine := fmt.Sprintf("%d %s\n", l.CurSequenceNumber, encoding)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Append to log file
	_, err = l.logWriter.Write([]byte(logLine))
	if err != nil {
		return err
	}
	l.CurSequenceNumber += 1

	return nil
}

func (l *Log) GetEntries() ([]*Entry, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	bytes, err := ioutil.ReadFile(l.Path)
	if err != nil {
		return nil, err
	}

	encodedEntries := strings.Split(string(bytes), "\n")
	res := make([]*Entry, len(encodedEntries))
	for i, encodedEntry := range encodedEntries {
		entry, err := DecodeLogEntry([]byte(encodedEntry))
		if err != nil {
			return nil, err
		}
		res[i] = entry
	}

	return res, nil
}

// Helper functions for dealing with the WAL

func DecodeLogEntry(entry []byte) (*Entry, error) {
	parts := strings.Split(string(entry), " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("there should be 2 parts in log line, got %d instead", len(parts))
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	unmarshalled := &Entry{}
	err = json.Unmarshal(decoded, unmarshalled)
	if err != nil {
		return nil, err
	}

	sequenceNumber, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	if sequenceNumber != unmarshalled.SequenceNumber {
		return nil, fmt.Errorf("sequence number and encoded sequence number do not match %d != %d", sequenceNumber, unmarshalled.SequenceNumber)
	}

	return unmarshalled, nil
}

type WALWriter struct {
	logPath string
	logFile *os.File
}

func (ww *WALWriter) Write(buf []byte) (int, error) {
	n, err := ww.logFile.Write(buf)
	if err != nil {
		return n, err
	}

	// Call fsync syscall to ensure log is immediately persisted to physical storage
	err = ww.logFile.Sync()
	if err != nil {
		return n, err
	}

	return n, nil
}

func NewWALWriter(path string) (*WALWriter, error) {
	// The WAL is kept as a persistently open append and write only file
	// TODO look into getting a system level lock on this file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WALWriter{
		logPath: path,
		logFile: file,
	}, nil
}
