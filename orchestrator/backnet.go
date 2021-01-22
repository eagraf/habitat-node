package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
)

type Backnet interface {
	StartProcess(backnet *entities.Backnet) (chan error, error)
}

type IPFSBacknet struct {
	backnet entities.Backnet
}

func (ib *IPFSBacknet) StartProcess(communityID entities.CommunityID) (chan error, error) {
	if ib.backnet.Type != entities.IPFS {
		return nil, errors.New("backnet should be of type IPFS")
	}

	// Make ipfs dir
	ipfsDir := filepath.Join(os.Getenv("IPFS_DIR"), string(communityID))
	err := os.MkdirAll(ipfsDir, 0700)
	if err != nil {
		return nil, err
	}

	env := []string{
		fmt.Sprintf("IPFS_PATH=%s", ipfsDir),
	}

	// Run ipfs init
	cmd := exec.Command("ipfs", "init")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)

	// Start ipfs daemon
	go func(errChan chan error) {
		cmd = exec.Command("ipfs", "daemon")
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			errChan <- err
		}
	}(errChan)

	return errChan, nil
}
