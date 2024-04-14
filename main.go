package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		exitUsage()
	}

	from := "en"
	queryStart := 1
	if strings.HasPrefix(os.Args[1], "from=") {
		if len(os.Args) < 3 {
			exitUsage()
		}
		from = strings.TrimPrefix(os.Args[1], "from=")
		queryStart = 2
	}

	query := strings.Join(os.Args[queryStart:], " ")
	to := []string{"en", "es", "fr", "ru", "lv", "lt"}
	to = slices.DeleteFunc(to, func(lang string) bool { return lang == from })
	slices.Sort(to)

	title, err := findTitle(from, query)
	if err != nil {
		log.Fatal(err)
	}

	links, err := getLangLinks(from, title)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %s\n", from, title) // "from" language is not included in lang links
	for _, lang := range to {
		i := slices.IndexFunc(links, func(ll LangLink) bool { return ll.Lang == lang })
		star := "???"
		if i != -1 {
			star = links[i].Star
		}
		fmt.Printf("%s: %s\n", lang, star)
	}
}

func findTitle(lang, query string) (title string, err error) {
	u, err := url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=query&format=json&list=search&srlimit=1")
	if err != nil {
		panic("failed to parse URL")
	}
	q := u.Query()
	q.Set("srsearch", query)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var search struct {
		Query struct {
			Searches []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}
	err = json.Unmarshal(respBytes, &search)
	if err != nil {
		return "", err
	}

	if len(search.Query.Searches) < 1 {
		return "", fmt.Errorf("No results for %s\n", query)
	}

	return search.Query.Searches[0].Title, nil
}

type LangLink struct {
	Lang string `json:"lang"`
	// LangName string `json:"langname"`
	Star string `json:"*"`
	// Url      string `json:"url"`
}

func getLangLinks(lang, title string) (langLinks []LangLink, err error) {
	u, err := url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=query&format=json&prop=langlinks&lllimit=max") // &llprop=langname|url
	if err != nil {
		panic("failed to parse URL")
	}
	q := u.Query()
	q.Set("titles", title)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var langs struct {
		Query struct {
			Pages map[string]struct {
				LangLinks []LangLink `json:"langlinks"`
			} `json:"pages"`
		} `json:"query"`
	}
	err = json.Unmarshal(respBytes, &langs)
	if err != nil {
		return nil, err
	}

	// return the first and only map entry
	for _, v := range langs.Query.Pages {
		return v.LangLinks, nil
	}
	return nil, errors.New("No results")
}

func exitUsage() {
	fmt.Println(`Usage: wiki-translate [from=es] multi-word search term`)
	os.Exit(0)
}
