package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
)

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Extract all http** links from a given webpage
func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls := [...]string{"https://allegro.pl/kategoria/obraz-i-grafika-karty-graficzne-257236?string=gtx+1070"}

	var itemsUrls []string //
	var categoryUrls []string //

	// Channels
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// We're done! Print the results...

	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url, _ := range foundUrls {
		var item = regexp.MustCompile(`i[0-9]+\.html$`)

		if item.MatchString(url) {
			itemsUrls = append(itemsUrls, url)
		}
		var category = regexp.MustCompile(`kategoria`)
		if category.MatchString(url) {
			categoryUrls = append(categoryUrls, url)
		}

	}
	fmt.Println( " Show item urls"  )
	i := 1
	for _, url := range itemsUrls {
		fmt.Println(" %v " + url,i)
		i++
	}
	i = 1
	fmt.Println( " Show category urls"  )
	for _, url := range categoryUrls {
		fmt.Println(" %v " + url,i)
		i++
	}






	close(chUrls)
}