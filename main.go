// ---------------------------
// FreshStart
// ---------------------------
// This will compile into a binary that can be run on a fresh system
// It will read the links.txt file and open the links in the browser
// ---------------------------

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Setsudan/FreshStart/internal/freshstart"
	"github.com/pkg/browser"
)

func main() {
	path := getInput()
	checkFileType(path)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	links, err := freshstart.ExtractLinkLines(file)
	if err != nil {
		log.Fatal(err)
	}

	openLinks(links)
	waitForInput()
}

func getInput() string {
	var path string
	println("Enter the path to the .txt file (same directory : 'example.txt' | parent directory : '../example.txt' | child directory : 'child/example.txt')")
	fmt.Scanln(&path)
	return path
}

func checkFileType(path string) {
	if !strings.HasSuffix(path, ".txt") {
		log.Fatal("File is not a .txt file")
	}
}

func openLinks(links []string) {
	for i, link := range links {
		if i == 0 {
			openLinkAndWait(link)
		} else {
			openLink(link)
		}
	}
}

func openLinkAndWait(link string) {
	_ = browser.OpenURL(link)
	var input string
	println("Press enter to continue")
	fmt.Scanln(&input)
}

func openLink(link string) {
	_ = browser.OpenURL(link)
}

func waitForInput() {
	var input string
	println("Press enter to exit")
	fmt.Scanln(&input)
}
