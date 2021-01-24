package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	config "github.com/ipfs/go-ipfs-config"
)

// IPFSConfig type allows us to attach our own methods to github.com/ipfs/go-ipfs-config config type
type IPFSConfig config.Config

// WriteConfig will write the config (pretty printed json) into specified file
func (c *IPFSConfig) WriteConfig(path string) error {
	// Marshal JSON indented
	buf, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	// Write config file
	err = ioutil.WriteFile(path, buf, 0600)
	if err != nil {
		return fmt.Errorf("Error writing config file: %s", err.Error())
	}

	return nil

}

// IPFSConfigBuilder allows for easy configuration of IPFS instances
type IPFSConfigBuilder struct {
	configuration *IPFSConfig
}

// NewIPFSConfigBuilder returns a new IPFSConfigBuilder, with a default identity
func NewIPFSConfigBuilder() (*IPFSConfigBuilder, error) {
	buffer := bytes.NewBuffer(make([]byte, 256))
	config, err := config.Init(buffer, 2048)
	if err != nil {
		return nil, err
	}
	cast := IPFSConfig(*config)

	return &IPFSConfigBuilder{
		configuration: &cast,
	}, nil
}

// SetIdentity overrides the newly generated identity provided by NewIPFSConfigBuilder
func (cb *IPFSConfigBuilder) SetIdentity(identity config.Identity) {
	cb.configuration.Identity = identity
}

// SetAddresses overrides the default ports used by IPFS:
// 4001: Swarm
// 5001: API
// 8080: Gateway
func (cb *IPFSConfigBuilder) SetAddresses(swarm, api, gateway int) {
	// TODO Validate port numbers
	addresses := config.Addresses{
		Swarm: []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", swarm),
			fmt.Sprintf("/ip6/::/tcp/%d", swarm),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", swarm),
			fmt.Sprintf("/ip6/::/udp/%d/quic", swarm),
		},
		API:        config.Strings{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", api)},
		Gateway:    config.Strings{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", gateway)},
		Announce:   make([]string, 0),
		NoAnnounce: make([]string, 0),
	}
	cb.configuration.Addresses = addresses
}

// Config returns the built configuration struct
func (cb *IPFSConfigBuilder) Config() *IPFSConfig {
	return cb.configuration
}
