package state

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eagraf/habitat-node/entities/transitions"
)

type Snapshot struct {
	DataB64        string                                     `json:"data"`
	Type           transitions.TransitionSubscriptionCategory `json:"type"`
	SequenceNumber uint64                                     `json:"sequence_number"`
	Timestamp      time.Time                                  `json:"timestamp"`
}

func WriteSnapshot(writer io.Writer, data interface{}, sequenceNumber uint64) error {
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

// Helper function to copy snapshot file to permanent version with timestamped name
func ArchiveSnapshotFile(path string, sequenceNumber int) error {
	oldFilePath := filepath.Join(path, "snapshot")
	newFilePath := filepath.Join(path, fmt.Sprintf("snapshot-%s-%s", strconv.Itoa(sequenceNumber), strconv.Itoa(int(time.Now().Unix()))))

	_, err := os.Stat(oldFilePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	oldFile, err := os.Open(oldFilePath)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, oldFile)
	if err != nil {
		return err
	}

	return nil
}
