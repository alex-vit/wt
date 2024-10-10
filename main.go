package main

import (
	"encoding/json"
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

// Finds the matching article and returns its title and URL.
// Uses opensearch API: https://www.mediawiki.org/wiki/API:Opensearch.
//
// Response is in the idiotic format of `[ string | []string ]`. Got the idea
// for how to parse it from https://gist.github.com/crgimenes/c3b8b4fcce8529e9201f83c8da134f32.
func findTitle(lang, query string) (title, titleUrl string, err error) {
	reqUrl := Must(url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=opensearch&format=json&redirects=resolve&limit=1"))
	q := reqUrl.Query()
	q.Set("search", query)
	reqUrl.RawQuery = q.Encode()

	resp, err := http.Get(reqUrl.String())
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// parsing response like:
	// ["aardvark",["Aardvark"],[""],["https://en.wikipedia.org/wiki/Aardvark"]]
	var idioticApi []any
	err = json.Unmarshal(respBytes, &idioticApi)
	if err != nil {
		return "", "", err
	}
	if len(idioticApi) != 4 {
		return "", "", fmt.Errorf("Expected 4 elements. Got: %#v", idioticApi)
	}

	titles, ok := idioticApi[1].([]any)
	if !ok {
		return "", "", fmt.Errorf("Expected a []any at [1]. Got: %#v", idioticApi[1])
	}
	if len(titles) == 0 {
		return "", "", fmt.Errorf(`No results for "%s"`, title)
	}
	title, ok = titles[0].(string)
	if !ok {
		return "", "", fmt.Errorf("Expected a string at [1][0]. Got: %#v", titles[0])
	}

	urls, ok := idioticApi[3].([]any)
	if !ok {
		return "", "", fmt.Errorf("Expected a []any at [3]. Got: %#v", idioticApi[3])
	}
	if len(urls) == 0 {
		return "", "", fmt.Errorf(`No results for "%s"`, title)
	}
	titleUrl, ok = urls[0].(string)
	if !ok {
		return "", "", fmt.Errorf("Expected a string at [3][0]. Got: %#v", urls[0])
	}

	return title, titleUrl, nil
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
	return nil, fmt.Errorf(`No results for "%s"`, title)
}

func exitUsage() {
	fmt.Println(`Usage: wiki-translate [from=es] multi-word search term`)
	os.Exit(0)
}

func Must[V any](value V, err error) V {
	if err != nil {
		panic(err)
	}
	return value
}
