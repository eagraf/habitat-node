package state

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"github.com/eagraf/habitat-node/entities"
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
	Transition *entities.TransitionWrapper

	SequenceNumber int64     `json:"sequence_number"`
	Committed      time.Time `json:"committed"`
}

// WriteAhead appends to the log file. This method should be called before anything else is done to process state
func (l *Log) WriteAhead(transition *entities.TransitionWrapper) error {
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
