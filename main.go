package main

import (
	"cmp"
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

	sourceLang := "en"
	queryStartIdx := 1
	if strings.HasPrefix(os.Args[1], "from=") {
		if len(os.Args) < 3 {
			exitUsage()
		}
		sourceLang = strings.TrimPrefix(os.Args[1], "from=")
		queryStartIdx = 2
	}

	query := strings.Join(os.Args[queryStartIdx:], " ")

	targetLangs := []string{"en", "es", "fr", "ru", "lv", "lt"}
	targetLangs = slices.DeleteFunc(targetLangs, func(lang string) bool { return lang == sourceLang })
	slices.Sort(targetLangs)

	title, url, err := findTitle(sourceLang, query)
	if err != nil {
		log.Fatal(err)
	}

	links, err := getLangLinks(sourceLang, title)
	if err != nil {
		log.Fatal(err)
	}

	// sort for binary search
	slices.SortFunc(links, func(a, b LangLink) int { return cmp.Compare(a.Lang, b.Lang) })

	fmt.Printf("%s: %-30s %s\n", sourceLang, title, url) // "from" language is not included in lang links
	for _, lang := range targetLangs {
		linkIdx, found := slices.BinarySearchFunc(links, lang, func(link LangLink, lang string) int {
			return cmp.Compare(link.Lang, lang)
		})
		if !found {
			fmt.Printf("%s: ???\n", lang)
			continue
		}
		star := links[linkIdx].Star
		if len(star) > 30 {
			star = star[:27] + "..."
		}
		fmt.Printf("%s: %-30s %s\n", lang, star, links[linkIdx].Url)
	}
}

// Finds the matching article and returns its title and URL.
// Uses opensearch API: https://www.mediawiki.org/wiki/API:Opensearch.
func findTitle(lang, query string) (title, titleUrl string, err error) {
	reqUrl := must(url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=opensearch&format=json&redirects=resolve&limit=1"))
	q := reqUrl.Query()
	q.Set("search", query)
	reqUrl.RawQuery = q.Encode()

	resp, err := http.Get(reqUrl.String())
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	loLoStr, err := listOfListsOfStrings(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("Failed to parse response: %w", err)
	}
	if len(loLoStr) != 4 || len(loLoStr[1]) == 0 || len(loLoStr[3]) == 0 {
		return "", "", fmt.Errorf("Malformed response. Expected a [4][1+]string, got: %v", loLoStr)
	}

	title = loLoStr[1][0]
	titleUrl = loLoStr[3][0]

	return title, titleUrl, nil
}

// Useful for parsing responses in the  format of `[ string | []string ]`.
// Idea from: https://gist.github.com/crgimenes/c3b8b4fcce8529e9201f83c8da134f32.
func listOfListsOfStrings(r io.Reader) ([][]string, error) {
	var anyList []any
	if err := json.NewDecoder(r).Decode(&anyList); err != nil {
		return nil, err
	}

	strLists := make([][]string, 0, len(anyList))
	for _, item := range anyList {
		switch obj := item.(type) {
		case string:
			strLists = append(strLists, []string{obj})
		case []any:
			strList := make([]string, 0, len(obj))
			for _, v := range obj {
				if str, ok := v.(string); ok {
					strList = append(strList, str)
				} else {
					return nil, fmt.Errorf("Expected a string but got %#v", v)
				}
			}
			strLists = append(strLists, strList)
		default:
			return nil, fmt.Errorf("Expected string or []any but got %v", obj)
		}
	}

	return strLists, nil
}

type LangLink struct {
	Lang string `json:"lang"`
	// LangName string `json:"langname"` // needs &llprop=langname
	Star string `json:"*"`
	Url  string `json:"url"` // needs &llprop=url
}

func getLangLinks(lang, title string) (langLinks []LangLink, err error) {
	u, err := url.Parse("https://" + lang + ".wikipedia.org/w/api.php?action=query&format=json&prop=langlinks&llprop=url&lllimit=max")
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

	var langs struct {
		Query struct {
			Pages map[string]struct {
				LangLinks []LangLink `json:"langlinks"`
			} `json:"pages"`
		} `json:"query"`
	}
	err = json.NewDecoder(resp.Body).Decode(&langs)
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
	fmt.Println(`Usage: wt [from=es] multi-word search term`)
	os.Exit(0)
}

func must[V any](value V, err error) V {
	if err != nil {
		panic(err)
	}
	return value
}
