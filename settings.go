package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/alex-vit/util"
)

const (
	dirName  string = "wt"
	filename string = "settings.json"
)

type Settings struct {
	TargetLanguages []string `json:"target_languages"`
	SourceLanguage  string   `json:"source_language"`
}

func (s *Settings) Normalize() {
	if len(s.TargetLanguages) == 0 {
		s.TargetLanguages = []string{"en", "es", "fr"}
	} else {
		s.TargetLanguages = slices.DeleteFunc(s.TargetLanguages, UnsupportedLanguage)
		slices.Sort(s.TargetLanguages)
		s.TargetLanguages = slices.Compact(s.TargetLanguages)
	}
	if s.SourceLanguage == "" {
		if slices.Contains(s.TargetLanguages, "en") {
			s.SourceLanguage = "en"
		} else {
			s.SourceLanguage = s.TargetLanguages[0]
		}
	}
	if i, found := slices.BinarySearch(s.TargetLanguages, s.SourceLanguage); !found {
		s.TargetLanguages = slices.Insert(s.TargetLanguages, i, s.SourceLanguage)
	}
}

func LoadSettings() *Settings {
	settings := &Settings{}

	file, err := os.Open(SettingsPath())
	if err == nil {
		defer file.Close()
		util.Must(0, json.NewDecoder(file).Decode(&settings))
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	settings.Normalize()
	return settings
}

func (s *Settings) Save() {
	util.Must(0, os.MkdirAll(settingsDir(), os.ModePerm))
	file := util.Must(os.Create(SettingsPath()))
	defer file.Close()

	s.Normalize()
	s.PrettyPrint(file)
}

func settingsDir() string {
	return filepath.Join(util.Must(os.UserConfigDir()), dirName)
}

func SettingsPath() string {
	return filepath.Join(settingsDir(), filename)
}

func (s *Settings) PrettyPrint(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	util.Must(0, enc.Encode(s))
}
