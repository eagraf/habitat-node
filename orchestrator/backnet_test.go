package main

import (
	"testing"

	"gotest.tools/assert"
)

/*
func TestCreateIPFSBacknetProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("TestIPFSStartProcess is a long test")
	}

	backnet := entities.InitBacknet(entities.IPFS)
	process := IPFSBacknet{
		communityID: "community_0",
		backnet:     *backnet,
	}

	_, err := process.StartProcess()
	if err != nil {
		t.Errorf(err.Error())
	}
}*/

// Checks that backnet impls fit the interface
func TestBacknetInterface(t *testing.T) {
	ipfsBacknet := interface{}(&IPFSBacknet{})
	_, ok := ipfsBacknet.(Backnet)
	assert.Equal(t, true, ok)
}
