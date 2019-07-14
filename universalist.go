package universalist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mgutz/ansi"
)

var defaultKeywords = []annotation{
	annotation{
		Text:     "TODO",
		Color:    "yellow",
		Priority: 1,
	},
	annotation{
		Text:     "FIXME",
		Color:    "cyan",
		Priority: 1,
	},
	annotation{
		Text:     "URGENT",
		Color:    "magenta",
		Priority: 1,
	},
	annotation{
		Text:     "BUG",
		Color:    "red",
		Priority: 1,
	},
}

// Universalister is a module Universalister
type Universalister struct {
	Path          string       `json:"path"`
	Keywords      []annotation `json:"keywords"`
	ExcludedFiles []string     `json:"excluded"`
	re            *regexp.Regexp
	list          list
	writer        io.Writer
}

type annotation struct {
	Text     string `json:"text"`
	Color    string `json:"color"`
	Priority int    `json:"priority"`
}

type list map[string][]locationInfo

type locationInfo struct {
	at          annotation
	instruction string
	filename    string
	Row         int
}

// Option is a type for building the module struct with different parameter
type Option func(*Universalister)

// New creates module struct with config path and given parameters
func New(configPath string, options ...Option) (*Universalister, error) {
	ul := &Universalister{
		Keywords: defaultKeywords,
		writer:   os.Stdout,
	}

	if configPath != "" {
		err := readConfig(ul, configPath)
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range options {
		opt(ul)
	}

	ul.re = makeRegexp(ul.Keywords)

	return ul, nil
}

// WithPath is an optional parameter that specifies the path that will be scanned
func WithPath(path string) Option {
	return func(ul *Universalister) {
		ul.Path = path
	}
}

// WithWriter is an optional parameter that specifies the writer that will be used to show the output
func WithWriter(writer io.Writer) Option {
	return func(ul *Universalister) {
		ul.writer = writer
	}
}

// Start starts collecting and listing all highlighted list
func (ul *Universalister) Start() error {
	err := ul.searchKeywords()
	if err != nil {
		return err
	}

	ul.showOutput()

	return nil
}

func (ul *Universalister) searchKeywords() error {
	list := make(list)

	err := filepath.Walk(ul.Path, func(path string, info os.FileInfo, err error) error {
		excluded, err := isExcluded(ul, path)
		if err != nil {
			return err
		}
		if !info.IsDir() && !excluded {
			localList, err := ul.searchKeywordsInFile(path)
			if err != nil {
				return err
			}

			for key, loc := range localList {
				list[key] = append(list[key], loc...)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	ul.list = list

	return nil
}

func (ul *Universalister) showOutput() {
	for key, locations := range ul.list {
		kw, ok := ul.getKeyword(key)
		if !ok {
			continue
		}

		fmt.Fprintln(ul.writer, ansi.Color(key, kw.Color))

		for _, loc := range locations {
			if loc.Row != 0 {
				fmt.Fprintln(ul.writer, loc.generateOutput())
			}
		}
	}
}

func (li locationInfo) generateOutput() string {
	return fmt.Sprintf("  - %s\t %s:%d\n", li.instruction, li.filename, li.Row)
}

func isExcluded(ul *Universalister, path string) (bool, error) {
	for _, pattern := range ul.ExcludedFiles {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func (ul *Universalister) searchKeywordsInFile(path string) (list, error) {
	list := make(list)
	row := 1

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	filename := stat.Name()

	defer f.Close()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		lineMatchedLocation, err := ul.getMatchedAnnotationLocations(sc.Text(), filename, row)
		if err != nil {
			row++
			continue
		}

		list[lineMatchedLocation.at.Text] = append(list[lineMatchedLocation.at.Text], lineMatchedLocation)

		row++
	}

	return list, nil
}

func (ul *Universalister) getMatchedAnnotationLocations(line, filename string, row int) (locationInfo, error) {
	key := ul.re.FindString(line)

	if key == "" {
		return locationInfo{}, nil
	}

	lineParts := strings.SplitAfter(line, ":")
	var instructions string

	if len(lineParts) > 1 {
		instructions = lineParts[1]
	}

	kw, ok := ul.getKeyword(key)
	if !ok {
		return locationInfo{}, fmt.Errorf("Keyword %s not found", key)
	}
	return locationInfo{
		at:          kw,
		instruction: instructions,
		filename:    filename,
		Row:         row,
	}, nil
}

func readConfig(ul *Universalister, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, ul)
	if err != nil {
		return err
	}

	return nil
}

func makeRegexp(keywords []annotation) *regexp.Regexp {
	pattern := "("
	var keys []string

	for _, keyword := range keywords {
		keys = append(keys, keyword.Text)
	}

	pattern += strings.Join(keys, "|") + ")"

	return regexp.MustCompile(pattern)
}

func (ul *Universalister) getKeyword(key string) (annotation, bool) {
	for _, kw := range ul.Keywords {
		if kw.Text == key {
			return kw, true
		}
	}

	return annotation{}, false
}
