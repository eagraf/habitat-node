package processes

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
	"golang.org/x/net/context"
)

type Backnet interface {
	ProcessID() ProcessID
	Configure(backnet *entities.Backnet) error
	StartProcess() (*Process, error)
}

type IPFSBacknet struct {
	communityID entities.CommunityID
	backnet     *entities.Backnet
	process     *Process // TODO refactor process out of here (we don't want a 2-way pointer

	ipfsDir    string
	configPath string
	config     *IPFSConfig
}

func InitIPFSBacknet(community *entities.Community, process *Process) (*IPFSBacknet, error) {
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

	process.CommunityID = community.ID

	return &IPFSBacknet{
		communityID: community.ID,
		backnet:     community.Backnet,
		ipfsDir:     ipfsDir,
		process:     process,
	}, nil
}

func (ib *IPFSBacknet) ProcessID() ProcessID {
	return ib.process.ID
}

func (ib *IPFSBacknet) Configure(newBacknet *entities.Backnet) error {
	if newBacknet.Type != entities.IPFS {
		return errors.New("backnet should be of type IPFS")
	}

	// TODO validate LocalBacknetConfig

	ipfsConfigPath := filepath.Join(os.Getenv("IPFS_DIR"), string(ib.communityID), "config")
	_, err := os.Stat(ipfsConfigPath)
	isNew := os.IsNotExist(err)

	ipfsDir := filepath.Join(os.Getenv("IPFS_DIR"), string(ib.communityID))
	err = os.MkdirAll(ipfsDir, 0700)
	if err != nil {
		return err
	}

	configDir := filepath.Join(os.Getenv("CONFIG_DIR"), string(ib.communityID))
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return err
	}
	ib.configPath = filepath.Join(configDir, "ipfs_config.json")

	// TODO factor out configuration logic and unit test

	var builder *IPFSConfigBuilder
	if isNew {
		builder, err = NewIPFSConfigBuilder()
		if err != nil {
			return err
		}
	} else {
		builder, err = NewIPFSConfigBuilderFromFile(filepath.Join(ipfsDir, "config"))
		if err != nil {
			return err
		}
	}

	config, err := buildConfig(builder, newBacknet)
	if err != nil {
		return err
	}

	err = config.WriteConfig(ib.configPath)
	if err != nil {
		return err
	}

	ib.config = config

	// If the backnet has not been initialized before, run ipfs init
	if isNew {

		env := []string{
			fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
		}

		// Run ipfs init
		ctx, _ := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, "ipfs", "init", ib.configPath)
		cmd.Env = env
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil

	} else {
		env := []string{
			fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
		}

		// Run ipfs config replace
		ctx, _ := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, "ipfs", "config", "replace", ib.configPath)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
		return nil
	}

	// Otherwise, make request to change config
	// TODO make sure daemon is restarted as well
}

func (ib *IPFSBacknet) StartProcess() (*Process, error) {
	if ib.backnet.Type != entities.IPFS {
		return nil, errors.New("backnet should be of type IPFS")
	}

	env := []string{
		fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
	}

	errChan := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ipfs", "daemon")
	cmd.Env = env

	// initialize process variables
	ib.process.context = ctx
	ib.process.cancel = cancel
	ib.process.errChan = errChan

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)

	// Start ipfs daemon
	err = cmd.Start()
	if err != nil {
		errChan <- err
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "Daemon is ready" {
			go func(cmd *exec.Cmd, errChan chan error) {
				//cmd.Stdout = os.Stdout
				//cmd.Stderr = os.Stderr
				err = cmd.Wait()
				if err != nil {
					errChan <- err
				}
			}(cmd, errChan)
			return ib.process, nil
		} else if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return nil, errors.New("cmd ended without receiving \"Daemon is ready\"")
}

func buildConfig(builder *IPFSConfigBuilder, backnet *entities.Backnet) (*IPFSConfig, error) {
	swarmPort, ok := backnet.Local.PortMap["swarm"]
	if !ok {
		return nil, errors.New("no swarm port included in port map")
	}
	apiPort, ok := backnet.Local.PortMap["api"]
	if !ok {
		return nil, errors.New("no api port included in port map")
	}
	gatewayPort, ok := backnet.Local.PortMap["gateway"]
	if !ok {
		return nil, errors.New("no gateway port included in port map")
	}

	builder.SetAddresses(swarmPort, apiPort, gatewayPort)
	config := builder.Config()
	return config, nil
}
