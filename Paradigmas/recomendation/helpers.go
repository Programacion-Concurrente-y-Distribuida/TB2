package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	colFechaPublicacion = 0
	colEntidad          = 2
	colDispositivo      = 3
	colNumero           = 4
	colSumilla          = 5
	colLink             = 6
	colFechaCorte       = 7
)

var lemmatizableCols = []int{colEntidad, colDispositivo, colSumilla}

type RowJob struct {
	Index int
	Row   []string
}

var accentMap = map[rune]rune{
	'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u', 'ü': 'u',
	'Á': 'a', 'É': 'e', 'Í': 'i', 'Ó': 'o', 'Ú': 'u', 'Ü': 'u',
}

func normalize(text string) string {
	lower := strings.ToLower(text)
	var b strings.Builder
	b.Grow(len(lower))
	for _, r := range lower {
		if rep, ok := accentMap[r]; ok {
			b.WriteRune(rep)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func loadStopwords(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]struct{})
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[normalize(line)] = struct{}{}
	}
	return out, scan.Err()
}

func loadLemmas(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string, 600_000)
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 64*1024), 1024*1024)
	for scan.Scan() {
		line := strings.TrimPrefix(scan.Text(), "\ufeff")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		lemma := normalize(strings.TrimSpace(parts[0]))
		form := normalize(strings.TrimSpace(parts[1]))
		if form == "" || lemma == "" {
			continue
		}
		if _, exists := out[form]; !exists {
			out[form] = lemma
		}
	}
	return out, scan.Err()
}

func tokenize(text string, stopwords map[string]struct{}) []string {
	norm := normalize(text)
	raw := strings.FieldsFunc(norm, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) < 2 {
			continue
		}
		if _, isStop := stopwords[t]; isStop {
			continue
		}
		out = append(out, t)
	}
	return out
}

func lemmatize(tokens []string, lemmas map[string]string) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		if l, ok := lemmas[t]; ok {
			out[i] = l
		} else {
			out[i] = t
		}
	}
	return out
}

func processText(text string, stopwords map[string]struct{}, lemmas map[string]string) string {
	return strings.Join(lemmatize(tokenize(text, stopwords), lemmas), " ")
}

func processRow(row []string, stopwords map[string]struct{}, lemmas map[string]string) {
	for _, idx := range lemmatizableCols {
		if idx < len(row) {
			row[idx] = processText(row[idx], stopwords, lemmas)
		}
	}
}

func withoutLinkColumn(row []string) []string {
	if colLink < 0 || colLink >= len(row) {
		return row
	}
	out := make([]string, 0, len(row)-1)
	out = append(out, row[:colLink]...)
	out = append(out, row[colLink+1:]...)
	return out
}

func mustOpenWriter(outPath string) (*os.File, error) {
	if dir := dirOf(outPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("crear directorio %s: %w", dir, err)
		}
	}
	return os.Create(outPath)
}

func dirOf(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return ""
}
