/*
Package doxiego communicates with a Doxie Go scanner.

The below example demonstrates how to use the API to communicate with a Doxie
scanner.

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

        jpeg.Encode(file, img, nil)

        // delete scans off the scanner
        ok, err := doxieGo.Delete("img_0001.jpg", "img_0002.jpg")
        if err != nil {
            //...
        } else if ok {
            fmt.Println("scans deleted.")
        }
    }
*/
package doxiego
