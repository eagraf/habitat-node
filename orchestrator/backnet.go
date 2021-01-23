package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
	"golang.org/x/net/context"
)

type Backnet interface {
	StartProcess(backnet *entities.Backnet) (*process, error)
}

type IPFSBacknet struct {
	backnet entities.Backnet
}

func (ib *IPFSBacknet) StartProcess(communityID entities.CommunityID) (*process, error) {
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
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ipfs", "init")
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

	process := &process{
		communityID: communityID,
		processType: processTypeBacknet,
		context:     ctx,
		cancel:      cancel,
		errChan:     errChan,
	}

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

	return process, nil
}
