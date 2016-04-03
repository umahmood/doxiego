package doxiego_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umahmood/doxiego"
)

// respFlags set by various tests, controls response of the test server
var respFlags struct {
	clientMode  bool
	emptyScans  bool
	noRecent    bool
	delNotFound bool
}

// byte representation on a jpeg image
var testScan = []byte{255, 216, 255, 224, 0, 16, 74, 70, 73, 70, 0, 1, 1, 1,
	0, 72, 0, 72, 0, 0, 255, 219, 0, 67, 0, 3, 2, 2, 2, 2, 2, 3, 2, 2, 2, 3,
	3, 3, 3, 4, 6, 4, 4, 4, 4, 4, 8, 6, 6, 5, 6, 9, 8, 10, 10, 9, 8, 9, 9, 10,
	12, 15, 12, 10, 11, 14, 11, 9, 9, 13, 17, 13, 14, 15, 16, 16, 17, 16, 10,
	12, 18, 19, 18, 16, 19, 15, 16, 16, 16, 255, 219, 0, 67, 1, 3, 3, 3, 4, 3,
	4, 8, 4, 4, 8, 16, 11, 9, 11, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 255, 192, 0, 17, 8, 0, 1, 0, 1, 3, 1, 34, 0, 2, 17, 1, 3, 17, 1, 255,
	196, 0, 21, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 255,
	196, 0, 20, 16, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 196,
	0, 20, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 255, 196, 0,
	20, 17, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 218, 0, 12,
	3, 1, 0, 2, 17, 3, 17, 0, 63, 0, 152, 0, 142, 126, 191, 255, 217}

func startTestServer() *httptest.Server {
	s := make(chan struct{})
	var ts *httptest.Server

	go func() {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			switch p {
			case "/hello.json":
				const r = string(`{ "model": "DX250",
                        "name": "Doxie_042D6A",
                        "firmwareWiFi": "1.29",
                        "hasPassword": false,
                        "MAC": "00:11:E5:04:2D:6A",
                        "mode": "AP",
                        "network": "",
                        "ip": ""}`)
				fmt.Fprintf(w, r)
			case "/hello_extra.json":
				const r = string(`{ "firmware": "0.26",
                        "connectedToExternalPower": true}`)
				fmt.Fprintf(w, r)
			case "/restart.json":
				w.WriteHeader(http.StatusNoContent)
			case "/scans.json":
				if respFlags.emptyScans {
					fmt.Fprintf(w, "[]")
				} else {
					const r = string(`[{
	                    "name":"/DOXIE/JPEG/IMG_0001.JPG",
	                    "size":241220,
	                    "modified":"2010-05-01 00:10:06"
	                    },
	                    {
	                    "name":"/DOXIE/JPEG/IMG_0002.JPG",
	                    "size":265085,
	                    "modified":"2010-05-01 00:09:26"
	                    },
	                    {
	                    "name":"/DOXIE/JPEG/IMG_0003.JPG",
	                    "size":273522,
	                    "modified":"2010-05-01 00:09:44"
	                    }]`)
					fmt.Fprintf(w, r)
				}
			case "/scans/recent.json":
				if respFlags.noRecent {
					w.WriteHeader(http.StatusNoContent)
				} else {
					const r = string(`{"path":"/DOXIE/JPEG/IMG_0003.JPG"}`)
					fmt.Fprintf(w, r)
				}
			case "/scans/DOXIE/JPEG/IMG_001.JPG":
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "image/jpeg")
				w.Write(testScan)
			case "/thumbnails/DOXIE/JPEG/IMG_001.JPG":
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "image/jpeg")
				w.Write(testScan)
			case "/scans/delete.json":
				if respFlags.delNotFound {
					w.WriteHeader(http.StatusForbidden)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		doxiego.APModeURL = ts.URL + "/"
		s <- struct{}{}
	}()
	_ = <-s
	return ts
}

func TestHelloAPMode(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	want := doxiego.Doxie{
		HasPassword:  false,
		Model:        "DX250",
		Name:         "Doxie_042D6A",
		FirmwareWiFi: "1.29",
		MAC:          "00:11:E5:04:2D:6A",
		Mode:         "AP",
		Network:      "",
		IP:           "",
		URL:          ts.URL + "/",
	}

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	if doxieGo.HasPassword != want.HasPassword {
		t.Errorf("hello: HasPassword want %t got %t",
			want.HasPassword, doxieGo.HasPassword)
	} else if doxieGo.Model != want.Model {
		t.Errorf("hello: Model want %s got %s",
			want.Model, doxieGo.Model)
	} else if doxieGo.Name != want.Name {
		t.Errorf("hello: Name want %s got %s",
			want.Name, doxieGo.Name)
	} else if doxieGo.FirmwareWiFi != want.FirmwareWiFi {
		t.Errorf("hello: WiFiFirmware want %s got %s",
			want.FirmwareWiFi, doxieGo.FirmwareWiFi)
	} else if doxieGo.MAC != want.MAC {
		t.Errorf("hello: MAC want %s got %s",
			want.MAC, doxieGo.MAC)
	} else if doxieGo.Mode != want.Mode {
		t.Errorf("hello: Mode want %s got %s",
			want.Mode, doxieGo.Mode)
	} else if doxieGo.Network != want.Network {
		t.Errorf("hello: Network want %s got %s",
			want.Network, doxieGo.Network)
	} else if doxieGo.IP != want.IP {
		t.Errorf("hello: IP want %s got %s",
			want.IP, doxieGo.IP)
	} else if doxieGo.URL != want.URL {
		t.Errorf("hello: URL want %s got %s",
			want.URL, doxieGo.URL)
	}
}

func TestHelloClientMode(t *testing.T) {
	fmt.Println("*** NOT IMPLEMENTED ***")
}

func TestPasswordField(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	wantPwd := "mypassword"

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	doxieGo.Password = "mypassword"

	got := doxieGo.Password

	if got != wantPwd {
		t.Errorf("set auth: invalid password want %s got %s", wantPwd, got)
	}
}

func TestScannerFirmware(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	want := "0.26"

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.ScannerFirmware()
	if err != nil {
		t.Errorf("%s", err)
	}

	if got != want {
		t.Errorf("scanner firmware: got %s want %s", got, want)
	}
}

func TestExternalPower(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	want := true

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.ExternalPower()
	if err != nil {
		t.Errorf("%s", err)
	}

	if got != want {
		t.Errorf("external power: got %t want %t", got, want)
	}
}

func TestRestart(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	if err := doxieGo.Restart(); err != nil {
		t.Errorf("%s", err)
	}
}

func TestScansWithResults(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	var want []doxiego.ScanItem

	want = append(want, doxiego.ScanItem{Name: "IMG_0001.JPG",
		Size:     241220,
		Modified: "2010-05-01 00:10:06"})

	want = append(want, doxiego.ScanItem{Name: "IMG_0002.JPG",
		Size:     265085,
		Modified: "2010-05-01 00:09:26"})

	want = append(want, doxiego.ScanItem{Name: "IMG_0003.JPG",
		Size:     273522,
		Modified: "2010-05-01 00:09:44"})

	scans, err := doxieGo.Scans()
	if err != nil {
		t.Errorf("%s", err)
	}

	for idx, got := range scans {
		if got.Name != want[idx].Name {
			t.Errorf("scan: Name got %s want %s", got.Name, want[idx].Name)
		} else if got.Size != want[idx].Size {
			t.Errorf("scan: Size got %d want %d", got.Size, want[idx].Size)
		} else if got.Modified != want[idx].Modified {
			t.Errorf("scan: Modified got %s want %s", got.Modified, want[idx].Modified)
		}
	}
}

func TestScansNoResults(t *testing.T) {
	ts := startTestServer()

	defer func() {
		respFlags.emptyScans = false
		ts.Close()
	}()

	respFlags.emptyScans = true

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	scans, err := doxieGo.Scans()
	if err != nil {
		t.Errorf("%s", err)
	}

	if len(scans) != 0 {
		t.Errorf("scan: slice length want %d got %d", 0, len(scans))
	}
}

func TestRecentWithResult(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	want := "IMG_0003.JPG"

	got, err := doxieGo.Recent()
	if err != nil {
		t.Errorf("%s", err)
	}

	if got != want {
		t.Errorf("recent: want %s got %s", want, got)
	}
}

func TestRecentNoResult(t *testing.T) {
	ts := startTestServer()
	defer func() {
		ts.Close()
		respFlags.noRecent = false
	}()

	respFlags.noRecent = true

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.Recent()
	if err != nil {
		t.Errorf("%s", err)
	}

	if got != "" {
		t.Errorf("recent: want %s got %s", "", got)
	}
}

func TestScanFound(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.Scan("IMG_001.JPG")
	if err != nil {
		t.Errorf("%s", err)
	}

	if got == nil {
		t.Errorf("scan: want non-nil image.Image got %v", got)
	}
}

func TestScanNotFound(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.Scan("IMG_BLAH.JPG")

	if err != doxiego.ErrScanNotFound {
		t.Errorf("%s", err)
	}

	if got != nil {
		t.Errorf("scan: want nil got %v", got)
	}
}

func TestThumbnailFound(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.Thumbnail("IMG_001.JPG")

	if err != nil {
		t.Errorf("%s", err)
	}

	if got == nil {
		t.Errorf("thumbnail: want non-nil image.Image got %v", got)
	}
}

func TestThumbnailNotFound(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	got, err := doxieGo.Thumbnail("IMG_BLAH.JPG")

	if err != doxiego.ErrNoThumbnail {
		t.Errorf("%s", err)
	}

	if got != nil {
		t.Errorf("thumbnail: scan value want nil got %v", got)
	}
}

func TestDeleteValidFiles(t *testing.T) {
	ts := startTestServer()
	defer ts.Close()

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	scansToDelete := []string{"IMG_001.JPG", "IMG_002.JPG"}

	got, err := doxieGo.Delete(scansToDelete...)
	if err != nil {
		t.Errorf("%s", err)
	}

	if got != true {
		t.Errorf("delete: want true got %t", got)
	}
}

func TestDeleteSomeInvalidFiles(t *testing.T) {
	ts := startTestServer()
	defer func() {
		ts.Close()
		respFlags.delNotFound = false
	}()

	respFlags.delNotFound = true

	doxieGo, err := doxiego.Hello()
	if err != nil {
		t.Errorf("%s", err)
	}

	scansToDelete := []string{"IMG_001.JPG", "IMG_002.JPG"}

	got, err := doxieGo.Delete(scansToDelete...)
	if err == nil {
		t.Errorf("delete: want non-nil err got nil")
	}

	if got != false {
		t.Errorf("delete: want false got %t", got)
	}
}
