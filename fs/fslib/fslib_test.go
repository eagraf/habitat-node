package fslib

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gotest.tools/assert"
)

/*
Commands:


fs ls community_0:/
fs mkdir community_0:/dir1/
fs mkdir community_0:/dir2/
fs ls community_0:/

fs pin community_0:/dir2/ check
fs pin community_0:/dir2/ unpin
fs pin community_0:/dir2/ check
fs pin community_0:/dir2/ pin

< create and write to file >

fs write community_0:/dir1/test.txt test.txt
fs ls community_0:/dir1/
fs copy community_0:/dir1/test.txt community_0:/dir2/test.txt
fs ls community_0:/dir2/
fs remove community_0:/dir2/test.txt
fs ls community_0:/dir2/
fs move community_0:/dir1/test.txt community_0:/dir2/
fs ls community_0:/dir1/
fs cat community_0:/dir2/test.txt


*/

func TestBasic(t *testing.T) {

	fs := &FSLibConfig{
		FStype: "IPFS",
		FSapi:  "127.0.0.1:6000",
	}

	comm := "community_0"

	res, err := fs.Ls(comm + ":/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "")

	res, err = fs.Mkdir(comm + ":/dir1/")
	res, err = fs.Mkdir(comm + ":/dir2/")

	res, err = fs.Ls(comm + ":/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "dir1, dir2")

	res, err = fs.Pin(comm+":/dir2/", "check")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "pinned")

	res, err = fs.Pin(comm+":/dir2/", "unpin")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn")

	res, err = fs.Pin(comm+":/dir2/", "check")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "not pinned")

	res, err = fs.Pin(comm+":/dir2/", "pin")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn")

	hw := "hello world!"
	f, err := os.Create("test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(hw)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	res, err = fs.Write("community_0:/dir1/test.txt", "test.txt")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	res, err = fs.Ls("community_0:/dir1/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "test.txt")

	res, err = fs.Copy("community_0:/dir1/test.txt", "community_0:/dir2/test.txt")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	res, err = fs.Ls("community_0:/dir2/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "test.txt")

	res, err = fs.Remove("community_0:/dir2/test.txt")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	res, err = fs.Ls("community_0:/dir2/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "")

	res, err = fs.Move("community_0:/dir1/test.txt", "community_0:/dir2/test.txt")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	res, err = fs.Ls("community_0:/dir1/")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "")

	res, err = fs.Cat("community_0:/dir2/test.txt")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	assert.Assert(t, strings.TrimSpace(res) == "hello world!")
}
