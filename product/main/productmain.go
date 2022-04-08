package main

import (
	"fmt"
	"github.com/josephspurrier/goversioninfo"
	"http-server.org/gazer/product/productinfo"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {

	makeSystem("../main/win32bundle.syso", productinfo.Name(), productinfo.Version(), "../main/favicon.ico", "../main/manifest.xml")

	compileInfo := fmt.Sprint("package productinfo\r\nconst BUILDTIME = \"" + time.Now().Format("2006-01-02 15:04:05.999") + "\"\r\n")
	err := ioutil.WriteFile("../product/productinfo/compileinfo.go", []byte(compileInfo), os.ModePerm)
	if err != nil {
		panic("Cannot write compile info file")
	}
}

func makeSystem(outputFile string, productName string, productVersion string, iconPath string, manifestPath string) {
	var vi goversioninfo.VersionInfo

	vi.ProductName = productName
	vi.StringFileInfo.ProductVersion = productVersion
	vi.IconPath = iconPath
	vi.ManifestPath = manifestPath

	vi.Build()
	vi.Walk()

	var archs []string
	//archs = []string{"386", "amd64"}
	//archs = []string{"386"} // 32-bit
	archs = []string{"amd64"} // 64-bit

	// Loop through each artchitecture.
	for _, item := range archs {
		// Create the file for the specified architecture.
		if err := vi.WriteSyso(outputFile, item); err != nil {
			log.Printf("Error writing syso: %v", err)
			os.Exit(3)
		}
	}
}
