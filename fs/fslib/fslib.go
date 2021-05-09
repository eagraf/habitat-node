package fslib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog/log"
)

type FSLibConfig struct {
	FStype string //ipfs, dat etc.
	FSapi  string // localhost:port for fs
}

// FSAPICall makes an HTTP GET request to the fs api running on local host
func FSAPICall(api string, httpPath string, args url.Values, file *os.File) (string, error) {

	url := url.URL{
		Scheme: "http",
		Host:   api,
		Path:   httpPath,
	}

	url.RawQuery = args.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("unable to make new HTTP Request %s", url.String()))
		return "", err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("unable to get response to %s", req.URL))
		return "", err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("")
		return "", err
	}

	log.Info().Msg("HTTP Response:\n" + string(bytes))
	return string(bytes), nil

}

func (fs FSLibConfig) Ls(path string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	return FSAPICall(fs.FSapi, "api/fs/ls", args, nil)
}

func (fs FSLibConfig) Write(path string, file string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	args.Set("file", file)
	return FSAPICall(fs.FSapi, "api/fs/write", args, nil)
}

func (fs FSLibConfig) Pin(path string, action string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	args.Set("action", action)
	return FSAPICall(fs.FSapi, "api/fs/pin", args, nil)
}

func (fs FSLibConfig) Remove(path string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	return FSAPICall(fs.FSapi, "api/fs/remove", args, nil)
}

func (fs FSLibConfig) Cat(path string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	return FSAPICall(fs.FSapi, "api/fs/cat", args, nil)
}

func (fs FSLibConfig) Move(old string, new string) (string, error) {
	args := url.Values{}
	args.Set("old", old)
	args.Set("new", new)
	return FSAPICall(fs.FSapi, "api/fs/move", args, nil)
}

func (fs FSLibConfig) Copy(old string, news string) (string, error) {
	args := url.Values{}
	args.Set("old", old)
	args.Set("new", news)
	return FSAPICall(fs.FSapi, "api/fs/copy", args, nil)
}

func (fs FSLibConfig) Mkdir(path string) (string, error) {
	args := url.Values{}
	args.Set("path", path)
	return FSAPICall(fs.FSapi, "api/fs/mkdir", args, nil)
}
