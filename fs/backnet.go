package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/eagraf/habitat-node/entities"
	"github.com/rs/zerolog/log"
)

// Backnet exposed to Filesystem (based off of Unix)
// most of these should just take in a string path = <community_id>:<path_to_file>
type Backnet interface {
	IsPinned(string) (bool, error)
	Pin(string) ([]byte, error)
	Unpin(string) ([]byte, error)

	ListFiles(string) ([]byte, error)
	Remove(string, bool) ([]byte, error) // bool = indicator of directory or file
	Cat(string) ([]byte, error)
	Write(string, *os.File) ([]byte, error)
	Move(string, string) ([]byte, error)
	Copy(string, string) ([]byte, error)
	MakeDir(string) ([]byte, error)
}

// IPFSBacknet implements these methods for an IPFS node
type IPFSBacknet struct {
	communityID entities.CommunityID
	backnet     entities.Backnet

	api string
}

// InitIPFSBacknet creates a filesystem-specific IPFS backnet
func InitIPFSBacknet(id entities.CommunityID, net entities.Backnet, port string) *IPFSBacknet {
	return &IPFSBacknet{
		communityID: id,
		backnet:     net,
		api:         port,
	}
}

// IPFSAPICall makes an HTTP API call and returns a string with the plain text
func IPFSAPICall(api string, httpPath string, args url.Values, file *os.File) (*http.Response, error) {

	url := url.URL{
		Scheme: "http",
		Host:   api,
		Path:   httpPath,
	}

	url.RawQuery = args.Encode()

	req, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to make new HTTP Request %s", url.String())
	}

	client := &http.Client{}

	if file != nil {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))

		if err != nil {
			log.Error().Err(err).Msg("")
		}

		io.Copy(part, file)
		writer.Close()
		req, err = http.NewRequest("POST", url.String(), body)

		if err != nil {
			log.Error().Err(err).Msg("")
		}

		req.Header.Add("Content-Type", writer.FormDataContentType())

	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response to %s", req.URL)
	}

	return res, nil

}

// FileInfoResponse is
type FileInfoResponse struct {
	Blocks         int    `json:"Blocks"`
	CumulativeSize uint64 `json:"CumulativeSize"`
	Hash           string `json:"Hash"`
	Local          bool   `json:"Local"`
	Size           uint64 `json:"Size"`
	SizeLocal      uint64 `json:"SizeLocal"`
	Type           string `json:"Type"`
	WithLocality   bool   `json:"WithLocality"`
}

func (net *IPFSBacknet) getHash(path string) (string, error) {

	argmap := map[string]string{"arg": path}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/stat",
		q,
		nil,
	)

	if err != nil {
		return "", err
	}

	// Read the response and return
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var resBodyJSON FileInfoResponse
	err = json.Unmarshal(resBody, &resBodyJSON)
	if err != nil {
		return "", err
	}

	log.Debug().Str("resBody", string(resBody)).Str("resBody hash", resBodyJSON.Hash).Msg("getHash")
	return resBodyJSON.Hash, nil
}

// Type is
type Type struct {
	Type string `json:"Type"`
}

// Keys is
type Keys struct {
	Pins map[string](Type) `json:"Keys"`
}

// PinListResponse is
// the json response should be this according to the API but it's not ....
type PinListResponse struct {
	PinLsList   Keys `json:"PinLsList"`
	PinLsObject map[string](string)
}

// PinListResponse2 is
// this is what the pinlist response acTually is
type PinListResponse2 struct {
	Message string `json:"Message"`
	Code    int    `json:"Code"`
	Type2   string `json:"Type"`
}

// IsPinned checks if a file is pinned on the users computer
func (net *IPFSBacknet) IsPinned(filepath string) (bool, error) {

	hash, err := net.getHash(filepath)
	if err != nil {
		return false, err
	}

	argmap := map[string]string{"arg": hash}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/pin/ls",
		q,
		nil,
	)

	if err != nil {
		return false, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var resBodyUnfoundJSON PinListResponse2
	err = json.Unmarshal(resBody, &resBodyUnfoundJSON)
	if err != nil {
		return false, err
	}
	if resBodyUnfoundJSON.Message != "" {
		// fmt.Println(resBodyUnfoundJSON.Message)
		return false, nil
	}

	var resBodyJSON Keys
	err = json.Unmarshal(resBody, &resBodyJSON)
	keys := resBodyJSON.Pins

	log.Debug().Str("resBody", string(resBody)).Str("hash", hash).Msg("IsPinned")
	if _, found := keys[hash]; found {
		return true, nil
	}

	return false, nil

}

type PinResponse struct {
	Pins []string `json:"Pins"`
}

// Pin implements pinning files locally for IPFS
func (net *IPFSBacknet) Pin(filepath string) ([]byte, error) {

	hash, err := net.getHash(filepath)
	if err != nil {
		return nil, err
	}

	argmap := map[string]string{"arg": hash}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/pin/add",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("response body", string(resBody)).Msg("Pin")

	var resBodyJSON PinResponse
	err = json.Unmarshal(resBody, &resBodyJSON)
	if err != nil {
		return nil, err
	}

	// bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(resBody)).Msg("Pin")
	return []byte(resBodyJSON.Pins[0]), nil
}

type UnPinResponse struct {
	Pins []string `json:"Pins"`
}

// Unpin implements unpinning a local file for IPFS
// Pin implements pinning files locally for IPFS
func (net *IPFSBacknet) Unpin(filepath string) ([]byte, error) {

	isPin, err := net.IsPinned(filepath)
	if err != nil {
		return nil, err
	}

	if isPin == false {
		return nil, errors.New("this file or directory has never been pinned")
	}

	hash, err := net.getHash(filepath)
	if err != nil {
		return nil, err
	}

	argmap := map[string]string{"arg": hash}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/pin/rm",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("response body", string(resBody)).Msg("UnPin")

	var resBodyJSON UnPinResponse
	err = json.Unmarshal(resBody, &resBodyJSON)
	if err != nil {
		return nil, err
	}

	// bytes, err := ioutil.ReadAll(res.Body)
	return []byte(resBodyJSON.Pins[0]), nil
}

type Entries struct {
	Hash string `json:"Hash"`
	Name string `json:"Name"`
	Size int64  `json:"Size"`
	Type int    `json:"Type"`
}

// PinListResponse is
// the json response should be this according to the API but it's not ....
type LsResponse struct {
	Entries []Entries `json:"Entries"`
}

// ListFiles implements ls for IPFSBacknets
func (net *IPFSBacknet) ListFiles(filepath string) ([]byte, error) {

	argmap := map[string]string{}
	if filepath != "" {
		argmap = map[string]string{"arg": filepath}
	}

	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/ls",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resBodyJSON LsResponse
	err = json.Unmarshal(resBody, &resBodyJSON)
	if err != nil {
		return nil, err
	}

	// bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(resBody)).Msg("List")
	var entries []string
	for _, e := range resBodyJSON.Entries {
		entries = append(entries, string(e.Name))
	}
	restr := strings.Join(entries, ", ")
	return []byte(restr), nil
}

// Remove implements rm for IPFSBacknets
func (net *IPFSBacknet) Remove(filepath string, isdir bool) ([]byte, error) {

	argmap := map[string]string{"arg": filepath}
	if isdir == true {
		argmap["force"] = "true"
	}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/rm",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(bytes)).Msg("Remove")
	return bytes, nil
}

// Cat implements cat for IFPSBacknets
func (net *IPFSBacknet) Cat(filepath string) ([]byte, error) {

	argmap := map[string]string{"arg": filepath}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"api/v0/files/read",
		q,
		nil,
	)

	buf := make([]byte, 0)
	if err != nil {
		return buf, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

// Write implements writing/updating files for IPFSBacknets
func (net *IPFSBacknet) Write(filepath string, f *os.File) ([]byte, error) {

	argmap := map[string]string{"arg": filepath, "create": "true", "parents": "true"}
	q := url.Values{}
	for arg, val := range argmap {
		q.Set(arg, val)
	}

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/write",
		q,
		f,
	)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(bytes)).Msg("Write")
	return bytes, nil
}

// Move implements mv for IPFSBacknets
// this needs to be tewsted bc online it says there are two arguments called arg for both
func (net *IPFSBacknet) Move(oldpath string, newpath string) ([]byte, error) {

	q := url.Values{}
	q.Add("arg", oldpath)
	q.Add("arg", newpath)

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/mv",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(bytes)).Msg("Move")
	return bytes, nil
}

// Copy implements cp for IPFSBacknets
// Also needs to be check/architecture needs to be modified
func (net *IPFSBacknet) Copy(oldpath string, newpath string) ([]byte, error) {

	q := url.Values{}
	q.Add("arg", oldpath)
	q.Add("arg", newpath)

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/cp",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(bytes)).Msg("Copy")
	return bytes, nil
}

// MakeDir implements mkdir for IPFSBacknets
func (net *IPFSBacknet) MakeDir(dirpath string) ([]byte, error) {

	q := url.Values{}
	q.Set("arg", dirpath)

	res, err := IPFSAPICall(
		net.api,
		"/api/v0/files/mkdir",
		q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	log.Debug().Str("response body", string(bytes)).Msg("MakeDir")
	return bytes, nil
}
