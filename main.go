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

	srcLang := "en"
	queryArgsStartIdx := 1
	if strings.HasPrefix(os.Args[1], "from=") {
		if len(os.Args) < 3 {
			exitUsage()
		}
		srcLang = strings.TrimPrefix(os.Args[1], "from=")
		queryArgsStartIdx = 2
	}

	query := strings.Join(os.Args[queryArgsStartIdx:], " ")
	targetLangs := []string{"en", "es", "fr", "ru", "lv", "lt"}
	targetLangs = slices.DeleteFunc(targetLangs, func(lang string) bool { return lang == srcLang })
	slices.Sort(targetLangs)

	title, url, err := findTitle(srcLang, query)
	if err != nil {
		log.Fatal(err)
	}

	links, err := getLangLinks(srcLang, title)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %-30s %s\n", srcLang, title, url) // "from" language is not included in lang links

	for _, lang := range targetLangs {
		linkIdx := slices.IndexFunc(links, func(ll LangLink) bool { return ll.Lang == lang })

		if linkIdx != -1 {
			star := links[linkIdx].Star
			if len(star) > 30 {
				star = star[:27] + "..."
			}
			url = links[linkIdx].Url
			fmt.Printf("%s: %-30s %s\n", lang, star, url)
		} else {
			fmt.Printf("%s: ???\n", lang)
		}
	}
}

func makeUrl(lang, query string) string {
	url, err := url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=query&format=json&list=search&srlimit=1")
	if err != nil {
		panic("failed to parse URL")
	}

	q := url.Query()
	q.Set("srsearch", query)
	url.RawQuery = q.Encode()

	return url.String()
}

func findTitle(lang, query string) (title, url string, err error) {
	url = makeUrl(lang, query)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	if len(search.Query.Searches) < 1 {
		return "", "", fmt.Errorf("No results for %s\n", query)
	}

	return search.Query.Searches[0].Title, url, nil
}

type LangLink struct {
	Lang string `json:"lang"`
	// LangName string `json:"langname"`
	Star string `json:"*"`
	Url  string `json:"url"`
}

func getLangLinks(lang, title string) (langLinks []LangLink, err error) {
	u, err := url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=query&format=json&prop=langlinks&llprop=url&lllimit=max") // &llprop=langname|url
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
