package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strings"

	"github.com/umahmood/doxiego"
)

var (
	help         bool
	scans        bool
	hello        bool
	delete       string
	getScans     bool
	getThumbanil string
	getScan      string
	auth         string
)

const emptyString = ""

func init() {
	flag.Usage = func() {
		printUsage()
	}

	flag.BoolVar(&help, "help", false, "Print this message and exit")
	flag.BoolVar(&hello, "hello", false, "Find Doxie Go on Wi-Fi network.")
	flag.BoolVar(&scans, "scans", false, "Display a list of all scans on the scanner.")
	flag.StringVar(&delete, "delete", emptyString, "Delete scans from the scanner.")
	flag.BoolVar(&getScans, "get-scans", false, "Download all scans on the scanner.")
	flag.StringVar(&getThumbanil, "get-thumbnail", emptyString, "Download a thumbnail from the scanner.")
	flag.StringVar(&getScan, "get-scan", emptyString, "Download a scan from the scanner.")
	flag.StringVar(&auth, "auth", emptyString, "Password to authenticate with the scanner.")

	flag.Parse()

	if flag.NFlag() == 0 {
		os.Exit(1)
	}

	if flag.NFlag() >= 2 && auth == emptyString {
		fmt.Println("to many command line flags, use '-help' for help.")
		os.Exit(1)
	}
}

func main() {
	if help {
		printUsage()
		os.Exit(0)
	}

	doxieGo, err := doxiego.Hello()
	checkError(err)

	if auth != emptyString {
		doxieGo.Password = auth
	}

	if hello {
		fmt.Println("Name:", doxieGo.Name)
		fmt.Println("Model:", doxieGo.Model)
		fmt.Println("Has Password:", doxieGo.HasPassword)
		fmt.Println("Wi-Fi Firmware:", doxieGo.FirmwareWiFi)
		fmt.Println("MAC:", doxieGo.MAC)
		if doxieGo.Mode == "AP" {
			fmt.Println("Mode:", doxieGo.Mode, "(Doxies own Wi-Fi network)")
		} else if doxieGo.Mode == "Client" {
			fmt.Println("Mode:", doxieGo.Mode, "(Doxie has joined existing Wi-Fi network)")
			fmt.Println("Network:", doxieGo.Network)
			fmt.Println("IP:", doxieGo.IP)
		}
		fmt.Println("URL:", doxieGo.URL)
	}

	if scans {
		items, err := doxieGo.Scans()
		checkError(err)
		for _, i := range items {
			fmt.Println("- name:", i.Name, "size:", i.Size, "modified:", i.Modified)
		}
	}

	if delete != emptyString {
		var dels []string
		x := strings.Split(delete, ",")
		for _, d := range x {
			if d != "" {
				dels = append(dels, strings.Trim(d, " "))
			}
		}
		_, err := doxieGo.Delete(dels...)
		checkError(err)
	}

	if getScans {
		items, err := doxieGo.Scans()
		checkError(err)
		for _, i := range items {
			img, err := doxieGo.Scan(i.Name)
			checkError(err)
			if err := saveImage(img, i.Name); err != nil {
				fmt.Println("error saving thumbnail", i.Name)
				fmt.Println(err)
			} else {
				fmt.Println("downloaded scan", i.Name)
			}
		}
	}

	if getThumbanil != emptyString {
		img, err := doxieGo.Thumbnail(getThumbanil)
		checkError(err)
		if err := saveImage(img, getThumbanil); err != nil {
			fmt.Println("error saving thumbnail", getThumbanil)
			fmt.Println(err)
		} else {
			fmt.Println("downloaded thumbnail", getThumbanil)
		}
	}

	if getScan != emptyString {
		img, err := doxieGo.Scan(getScan)
		checkError(err)
		if err := saveImage(img, getScan); err != nil {
			fmt.Println("error saving scan", getScan)
			fmt.Println(err)
		} else {
			fmt.Println("downloaded scan", getScan)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func saveImage(img image.Image, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return err
	}
	return nil
}

func printUsage() {
	fmt.Println(banner)
	fmt.Println(usage)
	fmt.Println(examples)
}

const banner = `
 ____   __  _  _  __  ____     ___   __
(    \ /  \( \/ )(  )(  __)   / __) /  \ 
 ) D ((  O ))  (  )(  ) _)   ( (_ \(  O )
(____/ \__/(_/\_)(__)(____)   \___/ \__/
`

const usage = `usage:

    -help           - Print this message and exit.
    -hello          - Find Doxie Go on Wi-Fi network.
    -scans          - Display a list of all scans on the scanner.
    -delete         - Delete scans from the scanner.
    -get-scans      - Download all scans on the scanner.
    -get-thumbnail  - Download a scan as a thumbnail from the scanner.
    -get-scan       - Download a scan from the scanner.
`

const examples = `example usage:

Find Doxie on the network:

$ doxiego -hello

Display a list of all scans:

$ doxiego -scans

Delete a list of scans (multiple scan names are comma separated):

$ doxiego -delete img_001.jpg,img_002.jpg

Download a scan as a thumbnail:

$ doxiego -get-thumbnail img_002.jpg

Download a scan:

$ doxiego -get-scan img_019.jpg

Download all scans:

$ doxiego -get-scans
`
