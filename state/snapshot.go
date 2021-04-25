package state

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/eagraf/habitat-node/entities/transitions"
)

type Snapshot struct {
	DataB64        string                                     `json:"data"`
	Type           transitions.TransitionSubscriptionCategory `json:"type"`
	SequenceNumber int                                        `json:"sequence_number"`
	Timestamp      time.Time                                  `json:"timestamp"`
}

func WriteSnapshot(writer io.Writer, data interface{}, sequenceNumber int) error {
	// Data is stored as base 64 encoded JSON
	marshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(marshalled)

	snapshot := &Snapshot{
		DataB64:        encoded,
		SequenceNumber: sequenceNumber,
		Timestamp:      time.Now(),
	}

	// Marshal snapshot into JSON and save to snapshot file
	marshalledSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	// Subtle but important. ioutil.WriteFile will overwrite and truncate the existing file
	// so the record of the old snapshot should be copied to a new file beforehand
	//	err = ioutil.WriteFile(path, marshalledSnapshot, 0744)
	_, err = writer.Write(marshalledSnapshot)
	if err != nil {
		return err
	}

	return nil
}

// ReadSnapshot reconsitutes state into the dest struct passed in, and returns the snapshot struct
func ReadSnapshot(reader io.Reader, dest interface{}) (*Snapshot, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var snapshot Snapshot
	err = json.Unmarshal(bytes, &snapshot)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(snapshot.DataB64)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decoded, dest)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}
