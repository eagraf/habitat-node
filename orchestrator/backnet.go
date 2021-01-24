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
	StartProcess() (*process, error)
}

type IPFSBacknet struct {
	communityID entities.CommunityID
	backnet     entities.Backnet
	process     process

	ipfsDir    string
	configPath string
	config     *IPFSConfig
}

func InitIPFSBacknet(community *entities.Community, swarmPort, apiPort, gatewayPort int) (*IPFSBacknet, error) {
	// Make ipfs dir
	ipfsDir := filepath.Join(os.Getenv("IPFS_DIR"), string(community.ID))
	err := os.MkdirAll(ipfsDir, 0700)
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(os.Getenv("CONFIG_DIR"), string(community.ID))
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return nil, err
	}

	builder, err := NewIPFSConfigBuilder()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "ipfs_config.json")
	builder.SetAddresses(swarmPort, apiPort, gatewayPort)
	config := builder.Config()
	config.WriteConfig(configPath)

	return &IPFSBacknet{
		communityID: community.ID,
		backnet:     community.Backnet,
		ipfsDir:     ipfsDir,
		configPath:  configPath,
		config:      config,
	}, nil
}

func (ib *IPFSBacknet) StartProcess() (*process, error) {
	if ib.backnet.Type != entities.IPFS {
		return nil, errors.New("backnet should be of type IPFS")
	}

	env := []string{
		fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
	}

	// Run ipfs init
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ipfs", "init", ib.configPath)
	cmd.Env = env
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)

	ib.process = process{
		communityID: ib.communityID,
		processType: processTypeBacknet,
		context:     ctx,
		cancel:      cancel,
		errChan:     errChan,
	}

	// Start ipfs daemon
	go func(errChan chan error) {
		cmd = exec.Command("ipfs", "daemon")
		cmd.Env = env
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			errChan <- err
		}
		err = cmd.Wait()
		if err != nil {
			errChan <- err
		}
	}(errChan)

	return &ib.process, nil
}
