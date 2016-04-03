package doxiego

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const doxieInternalPath = "/DOXIE/JPEG/"

var (
	// ErrHTTPRequest error when making a http request to the scanner
	ErrHTTPRequest error
	// ErrDoxieNotFound error when the scanner is not reachable
	ErrDoxieNotFound = errors.New("doxie: scanner not found on Wi-Fi network")
	// ErrScanNotFound error when scans.json endpoint returns an empty body
	ErrScanNotFound = errors.New("doxie: scan(s) not found scanners memory may be busy")
	// ErrDeletingScan error when delete.json endpoint cannot delete a scan
	ErrDeletingScan = errors.New("doxie: error deleting scan(s)")
	// ErrDownloadingScan request for scan returns no data
	ErrDownloadingScan = errors.New("doxie: error downloading scan")
	// ErrNoThumbnail thumbnail has not yet been generated.
	ErrNoThumbnail = errors.New("doxie: thumbnail no yet generated")
)

var (
	// APModeURL URL of scanner when it creates its own network
	APModeURL = "http://192.168.1.100:8080/"
	// StaticIP of the scanner when it joins client network
	StaticIP string
)

// Doxie represents a Doxie scanner instance
type Doxie struct {
	// Has password been set to authenticate API access.
	HasPassword bool
	// Scanner Model
	Model string
	// Name of the scanner, defaults to the form "Doxie_XXXXXX"
	Name string
	// FirmwareWiFi version
	FirmwareWiFi string
	// MAC address of the scanner
	MAC string
	// Mode signals if the scanner is in AP or Client mode
	Mode string
	// If in client mode, the name of the network joined
	Network string
	// If in client mode, the IP of the network joined
	IP string
	// URL of the Doxie API
	URL string
	// Scanner password base64 encoded
	Password string
}

// ScanItem list of scans in the scanners memory
type ScanItem struct {
	Name     string
	Size     int
	Modified string
}

// helloExtra additional status values from the doxie scanner
type helloExtra struct {
	// Scanners firmware version
	Firmware string
	// Scanners power source, true if AC false if battery power
	ConnectedToExternalPower bool
}

// response wraps a response from the doxie scanner
type response struct {
	statusCode int
	data       []byte
	err        error
}

// Hello returns status information for the scanner, firmware, network mode, and
// password configuration. Accessing this command does not require a password if
// one has been set. The values returned depend on whether the scanner is creating
// its own network or joining an existing network.
func Hello() (*Doxie, error) {
	findDoxieOnAPNetwork := func(url string, chd chan *Doxie, che chan error) {
		r := httpGetRequest(url+"hello.json", "")

		if r.err != nil {
			che <- r.err
			return
		}

		if r.statusCode != http.StatusOK {
			ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
			che <- ErrHTTPRequest
			return
		}

		var dox Doxie

		err := json.Unmarshal(r.data, &dox)
		if err != nil {
			che <- err
			return
		}

		dox.URL = APModeURL

		chd <- &dox
	}

	var dox *Doxie
	var err error

	chDox := make(chan *Doxie)
	chErr := make(chan error)

	// Find Doxie on the network it creates - 'AP' mode
	go findDoxieOnAPNetwork(APModeURL, chDox, chErr)

	// TODO: Find Doxie on the network it joins - 'Client' mode
	// go findDoxieOnClientNetwork(StaticIP, chDox, chErr)

	select {
	case dox = <-chDox:
	case err = <-chErr:
	}

	return dox, err
}

// ScannerFirmware the scanners firmware version.
func (d *Doxie) ScannerFirmware() (string, error) {
	r := httpGetRequest(d.URL+"hello_extra.json", "")

	if r.err != nil {
		return "", r.err
	}

	if r.statusCode != http.StatusOK {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return "", ErrHTTPRequest
	}

	var extra helloExtra

	err := json.Unmarshal(r.data, &extra)
	if err != nil {
		return "", err
	}

	return extra.Firmware, nil
}

// ExternalPower indicates whether the scanner is connected to its AC adapter
// versus running on battery power. This value is not cached, so it immediately
// reflects any state changes.
func (d *Doxie) ExternalPower() (bool, error) {
	r := httpGetRequest(d.URL+"hello_extra.json", "")

	if r.err != nil {
		return false, r.err
	}

	if r.statusCode != http.StatusOK {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return false, ErrHTTPRequest
	}

	var extra helloExtra

	err := json.Unmarshal(r.data, &extra)
	if err != nil {
		return false, err
	}

	return extra.ConnectedToExternalPower, nil
}

// Restart restarts the scanner's Wi-Fi system.
func (d *Doxie) Restart() error {
	r := httpGetRequest(d.URL+"restart.json", d.Password)

	if r.err != nil {
		return r.err
	}

	// DoxieGo returns http 204 No Content and then restarts the scanner's Wi-Fi
	// system. The scanner's status light blinks blue during the restart.
	if r.statusCode != http.StatusNoContent {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return ErrHTTPRequest
	}

	return nil
}

// Scans returns an array of all scans currently in the scanners memory. After
// scanning a document, the scan will available several seconds later. Calling
// this function immediately after scanning something may return a blank result,
// even if there are other scans on the scanner, due to the scanner's memory
// being in use. Consider retrying if len(ScanItems) is zero.
func (d *Doxie) Scans() ([]ScanItem, error) {
	r := httpGetRequest(d.URL+"scans.json", d.Password)

	if r.err != nil {
		return nil, r.err
	}

	if r.statusCode != http.StatusOK {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return nil, ErrHTTPRequest
	}

	var items []ScanItem

	// no data sent from scanner
	if len(r.data) == 0 {
		return nil, ErrScanNotFound
	}

	err := json.Unmarshal(r.data, &items)
	if err != nil {
		return nil, err
	}

	for idx, i := range items {
		items[idx].Name = path.Base(i.Name)
	}

	return items, nil
}

// Recent returns the last scan if available, if there is no recent scan
// available, an empty string is returned.
func (d *Doxie) Recent() (string, error) {
	r := httpGetRequest(d.URL+"scans/recent.json", d.Password)

	if r.err != nil {
		return "", r.err
	}

	if r.statusCode == http.StatusNoContent {
		return "", nil
	} else if r.statusCode != http.StatusOK {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return "", ErrHTTPRequest
	}

	var recent map[string]string

	err := json.Unmarshal(r.data, &recent)
	if err != nil {
		return "", err
	}

	_, ok := recent["path"]
	if !ok && len(recent) < 1 {
		return "", nil
	}

	return path.Base(recent["path"]), nil
}

// Scan gets a scanned item by name.
func (d *Doxie) Scan(name string) (image.Image, error) {
	return getScanHelper(d.URL, "scans", name, d.Password)
}

// Thumbnail gets a 240x240 thumbnail of the scan. Returns error ErrNoThumbnail
// if the thumbnail has not yet been generated, retrying after a delay is
// recommended to handle such cases.
func (d *Doxie) Thumbnail(name string) (image.Image, error) {
	img, err := getScanHelper(d.URL, "thumbnails", name, d.Password)
	if err == ErrScanNotFound {
		return img, ErrNoThumbnail
	}
	return img, err
}

// Delete deletes multiple scans in a single operation.
func (d *Doxie) Delete(items ...string) (bool, error) {
	var body string
	for idx, s := range items {
		if idx == len(items)-1 {
			body = body + strconv.Quote(doxieInternalPath+strings.ToUpper(s))
		} else {
			body = body + strconv.Quote(doxieInternalPath+strings.ToUpper(s)) + ","
		}
	}

	buf := bytes.NewBufferString("[" + body + "]")

	var url string
	if d.Password != "" {
		url = addAuthToURL(d.URL, d.Password)
	} else {
		url = d.URL
	}

	resp, err := http.Post(url+"scans/delete.json", "application/json", buf)
	if err != nil {
		return false, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	// scanner returns 204 if successful
	if resp.StatusCode != http.StatusNoContent {
		return false, ErrDeletingScan
	}

	return true, nil
}

// addAuthToURL inserts username:password into a URL.
func addAuthToURL(url, password string) string {
	if strings.HasPrefix(url, "http://") {
		url = fmt.Sprintf("http://doxie:%s@%s", password, url[7:])
	} else {
		// assumes d.URL is in form 192.168.1.100:8080
		url = fmt.Sprintf("http://doxie:%s@%s", password, url)
	}
	return url
}

// getScanHelper helper function retrieves a jpeg scan from the scanner.
func getScanHelper(url, path, name, password string) (image.Image, error) {
	r := httpGetRequest(url+path+doxieInternalPath+strings.ToUpper(name), password)

	if r.err != nil {
		return nil, r.err
	}

	// scanner returns 404 when scan can not be found.
	if r.statusCode == http.StatusNotFound {
		return nil, ErrScanNotFound
	} else if r.statusCode != http.StatusOK {
		ErrHTTPRequest = errors.New("doxie: request error http " + strconv.Itoa(r.statusCode))
		return nil, ErrHTTPRequest
	}

	if len(r.data) == 0 {
		return nil, ErrDownloadingScan
	}

	var b bytes.Buffer
	b.Write(r.data)

	img, err := jpeg.Decode(&b)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// httpGetRequest makes a request to a HTTP endpoint
func httpGetRequest(url, password string) *response {
	ch := make(chan *response)

	go func() {
		if password != "" {
			url = addAuthToURL(url, password)
		}
		resp, err := http.Get(url)

		if err != nil {
			ch <- &response{data: nil, err: err}
			return
		}

		if resp != nil {
			defer resp.Body.Close()
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ch <- &response{data: nil, err: err}
			return
		}

		ch <- &response{statusCode: resp.StatusCode, data: body, err: err}
	}()

	var resp *response
	select {
	case r := <-ch:
		resp = r
	case <-time.After(5 * time.Second):
		resp = &response{statusCode: http.StatusNotFound,
			data: nil,
			err:  ErrDoxieNotFound,
		}
	}

	return resp
}
