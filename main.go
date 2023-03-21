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

	"github.com/pkg/browser"
)

// Array of links
var links []string

func appendLink(link string) {
	links = append(links, link)
}

func main() {
	// Read the links.txt file
	file, err := os.Open("links.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
    		line := scanner.Text()
    		if !strings.HasPrefix(line, "#") {
        		appendLink(line)
    		}
	}


	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Open the links in the browser
	for _, link := range links {
		// After the first link is opened wait until user input
		if link == links[0] {
			browser.OpenURL(link)
			var input string
			println("Press enter to continue")
			fmt.Scanln(&input)
		} else {
			browser.OpenURL(link)
		}
	}

	// Wait for user input
	var input string
	println("Press enter to exit")
	fmt.Scanln(&input)
}
