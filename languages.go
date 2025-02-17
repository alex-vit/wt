package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"strings"
)

//go:embed wpcodes.txt
var wpcodes []byte
var supportedLanguageCodes map[string]struct{} = make(map[string]struct{}, 343)

func init() {
	scanner := bufio.NewScanner(bytes.NewReader(wpcodes))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		// skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		supportedLanguageCodes[line] = struct{}{}
	}
}

func IsSupportedLanguage(code string) bool {
	_, found := supportedLanguageCodes[code]
	return found
}

func UnsupportedLanguage(code string) bool {
	return !IsSupportedLanguage(code)
}
