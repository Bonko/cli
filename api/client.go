package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/stormforger/cli/buildinfo"
)

type dialer func(network, addr string) (net.Conn, error)

func newDialer(fingerprint []byte, skipCAVerification bool) dialer {
	return func(network, addr string) (net.Conn, error) {
		c, err := tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: skipCAVerification})
		if err != nil {
			return c, err
		}
		connstate := c.ConnectionState()
		keyPinValid := false
		for _, peercert := range connstate.PeerCertificates {
			der, err := x509.MarshalPKIXPublicKey(peercert.PublicKey)
			hash := sha256.Sum256(der)
			// log.Println(peercert.Issuer)
			// log.Printf("%#v\n\n", hash)
			if err != nil {
				log.Fatal(err)
			}
			if bytes.Compare(hash[0:], fingerprint) == 0 {
				// log.Println("Pinned Key found")
				keyPinValid = true
			}
		}
		if keyPinValid == false {
			log.Fatal("TLS Public Key could not be verified!")
		}
		return c, nil
	}
}

func defaultHTTPClient() *http.Client {
	fingerprint := []byte{0x5b, 0x15, 0x6c, 0xda, 0x7b, 0xc3, 0xd, 0x8b, 0xe8, 0x88, 0x57, 0x75, 0xbc, 0x30, 0xc1, 0x84, 0x18, 0x75, 0x6f, 0x2d, 0x3b, 0x81, 0x91, 0xff, 0x34, 0x10, 0xda, 0x13, 0x4a, 0x83, 0x23, 0x9d}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		DialTLS:            newDialer(fingerprint, false),
	}

	return &http.Client{
		Transport: tr,
		Timeout:   2 * time.Minute,
	}
}

// Client represents the API client
type Client struct {
	HTTPClient  *http.Client
	APIEndpoint string
	JWT         string
	UserAgent   string
}

// NewClient returns a new initialized API client
func NewClient(apiEndpoint string, jwtToken string) *Client {
	return &Client{
		HTTPClient:  defaultHTTPClient(),
		APIEndpoint: apiEndpoint,
		JWT:         jwtToken,
		UserAgent:   fmt.Sprintf("StormForger-CLI/%v (%v; +https://github.com/stormforger/cli)", buildinfo.BuildInfo.Version, buildinfo.BuildInfo.Commit),
	}
}

// Ping performs an authenticated ping to check if the API
// is working and the user is properly authenticated.
//
// FIXME would be nice to return a struct
//       where we see the status and in case of
//       success also the email address of the
//       authenticated user (useful) to check
//       if we are authenticated as the correct user
func (c *Client) Ping() (bool, error) {
	req, err := http.NewRequest("GET", c.APIEndpoint+"/authenticated_ping", nil)

	c.setUserAgent(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	return (resp.StatusCode == 200), errors.New("could not perform authenticated ping")
}

func newPatchRequest(uri string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func (c *Client) addDefaultHeaders(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+c.JWT)
	c.setUserAgent(request)
}

func (c *Client) setUserAgent(request *http.Request) {
	request.Header.Set("User-Agent", c.UserAgent)
}

func (c *Client) doRequestRaw(request *http.Request) (*http.Response, error) {
	c.addDefaultHeaders(request)

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) doRequest(request *http.Request) ([]byte, error) {
	response, err := c.doRequestRaw(request)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return body, nil
}
