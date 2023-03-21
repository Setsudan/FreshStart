// ---------------------------
// FreshStart
// ---------------------------
// This will compile into a binary that can be run on a fresh system
// It will read the links.txt file and open the links in the browser
// ---------------------------

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/browser"
)

// Array of links
var links []string

func appendLink(link string) {
	links = append(links, link)
}

func main() {
	path := getInput()
	file := openFile(path)
	defer file.Close()
	checkFileType(path)

	links, err := extractLinks(file)
	if err != nil {
		log.Fatal(err)
	}

	openLinks(links)
	waitForInput()
}

func getInput() string {
	var path string
	println("Enter the path to the .txt file")
	fmt.Scanln(&path)
	return path
}

func openFile(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func checkFileType(path string) {
	if !strings.HasSuffix(path, ".txt") {
		log.Fatal("File is not a .txt file")
	}
}

func extractLinks(file *os.File) ([]string, error) {
	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			links = append(links, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return links, nil
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
	browser.OpenURL(link)
	var input string
	println("Press enter to continue")
	fmt.Scanln(&input)
}

func openLink(link string) {
	browser.OpenURL(link)
}

func waitForInput() {
	var input string
	println("Press enter to exit")
	fmt.Scanln(&input)
}
