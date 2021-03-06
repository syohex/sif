package sif

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Match struct {
	Line int
	Text string
}

type FileMatched struct {
	Name    string
	Matches []Match
}

type Options struct {
	CaseInsensitive bool
}

type Sif struct {
	pattern *regexp.Regexp
	options Options
}

func (s *Sif) Scan(path string) ([]*FileMatched, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if f.IsDir() {
		return s.ScanDir(path)
	}

	fm, err := s.ScanFile(path)
	if fm != nil {
		return []*FileMatched{fm}, err
	}
	return nil, err
}

func (s *Sif) ScanDir(dir string) ([]*FileMatched, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filesMatched := make([]*FileMatched, 0)
	for _, f := range files {
		path := filepath.Join(dir, f.Name())
		if !f.IsDir() {
			fm, err := s.ScanFile(path)
			if err != nil {
				return nil, err
			}
			if fm != nil {
				filesMatched = append(filesMatched, fm)
			}
		} else if !ignoreDirs.MatchString(f.Name()) {
			fs, err := s.ScanDir(path)
			if err != nil {
				return nil, err
			}
			filesMatched = append(filesMatched, fs...)
		}
	}

	return filesMatched, nil
}

func (s *Sif) ScanFile(path string) (*FileMatched, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if b, err := isBinary(file); b || err != nil {
		return nil, err
	}

	line := 1
	matches := make([]Match, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if s.pattern.MatchString(text) {
			text := s.pattern.ReplaceAllStringFunc(text, bgYellow)
			matches = append(matches, Match{line, text})
		}
		line++
	}

	if len(matches) > 0 {
		return &FileMatched{path, matches}, nil
	}

	return nil, nil
}

func New(pattern string, options Options) *Sif {
	if options.CaseInsensitive {
		pattern = fmt.Sprintf("(?i)%s", pattern)
	}
	p := regexp.MustCompile(pattern)
	return &Sif{p, options}
}
