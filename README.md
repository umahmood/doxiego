# DoxieGo

DoxieGo is command line tool and Go library to communicate with a [Doxie Go Wi-Fi scanner](http://www.getdoxie.com/product/doxie-go/).

Note: This project is not affiliated with [Apparent](http://www.itsapparent.com/products/index.html) or [Doxie](http://www.getdoxie.com/).

Note: Currently the code only works when the Doxie is in AP mode (your computer joins Doxies built in wi-fi network).

# Installation

> go get github.com/umahmood/doxiego <br/>
> cd $GOPATH/src/github.com/umahmood/doxiego <br/>
> go test ./... <br/>

# Usage from the command line

Find Doxie on the network:

> $ doxiego -hello <br/>
Name: Doxie_0591E0 <br/>
Model: DX250 <br/>
Has Password: false <br/>
Wi-Fi Firmware: 1.29 <br/>
MAC: omitted :) <br/>
Mode: AP (Doxies own Wi-Fi network) <br/>
URL: http://192.168.1.100:8080/ <br/>

Display a list of all scans:

> $ doxiego -scans <br/>
- name: IMG_0002.JPG size: 959458 modified: 2010-05-01 00:03:26 <br/>
- name: IMG_0003.JPG size: 941949 modified: 2010-05-01 00:06:44 <br/>

Delete a list of scans (multiple scan names are comma separated):

> $ doxiego -delete img_0002.jpg,img_0003.jpg

Download a scan as a thumbnail:

> $ doxiego -get-thumbnail img_0002.jpg <br/>
downloaded thumbnail img_0002.jpg  

Download a scan:

> $ doxiego -get-scan img_0003.jpg <br/>
downloaded scan img_0003.jpg  

Download all scans:

> $ doxiego -get-scans <br/>
downloaded scan IMG_0002.JPG <br/>
downloaded scan IMG_0003.JPG <br/>

For help:

> $ doxiego -help <br/>
...

# Usage from the API:

See GoDoc for full capability of the API.

    package main

    import (
        "fmt"

        "github.com/umahmood/doxiego"
    )

    func main() {
        doxieGo, err := doxiego.Hello()
        if err != nil {
            //...
        }

        fmt.Println("Doxie name", doxieGo.Name)

        // if the scanner has a password set, fill in the password field
        doxieGo.Password = "mypassword"

        // get a list of scanned items on the scanner
        items, err := doxieGo.Scans()
        if err != nil {
            //...
        }
        
        for _, s := range items {
            fmt.Println("name:", s.Name, "size:", s.Size, "modified:", s.Modified)
        }

        // download a scan
        img, err := doxieGo.Scan("img_0001.jpg")
        if err != nil {
            //...
        }
        //...
        jpeg.Encode(file, img, nil)

        // delete scans off the scanner
        ok, err := doxieGo.Delete("img_0001.jpg", "img_0002.jpg")
        if err != nil {
            //...
        } else if ok {
            fmt.Println("scans deleted.")
        }
    }

# Documentation

> http://godoc.org/github.com/umahmood/doxiego

# Testing

- The codebase has been tested with a single Doxie Go Wi-Fi scanner. The API could be used to communicate with multiple Doxie Go Wi-Fi scanners.

# License

See the [LICENSE](LICENSE.md) file for license rights and limitations (MIT).
