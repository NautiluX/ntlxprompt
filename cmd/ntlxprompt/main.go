package main

import (
	"os"
	"path/filepath"
	"fmt"
	"strings"
	"github.com/go-git/go-git/v5"
	"golang.org/x/term"
)

func getGitDir() string {
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return gitPath
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir { // Reached root
			return ""
		}
		dir = parentDir
	}
}

var (
	red   = color("\033[31m%s\033[0m")
	green = color("\033[32m%s\033[0m")
	boldgreen = color("\033[1;32m%s\033[0m")
	boldcyan  = color("\033[1;36m%s\033[0m")
	boldwhite = color("\033[1;37m%s\033[0m")
	boldpurple = color("\033[1;35m%s\033[0m")
)

func color(s string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(s, fmt.Sprint(args...))
	}
}

const (
	promptSym = ""
)

var (
	cwd string
	home string
)

func trimPath(cwd, home string) string {
	width, _, err := term.GetSize(0)
  if err != nil {
		width = 80
  }

	path := strings.Replace(cwd, home, "~", 1)
	if len(path) <= width/4 {
		return path;
	}
	items := strings.Split(path, "/")
	truncItems := []string{}
	for i, item := range items {
		if i != (len(items) - 1) && i != 0 {
			truncItems = append(truncItems, item[:1])
			continue
		}
		truncItems = append(truncItems, item)
	}
	if len(truncItems) > 0 && truncItems[0] != "~" {
		truncItems[0] = "/"
	}
	return filepath.Join(truncItems...)
}

func addDirSeg(segments []string) []string {
	return append(segments,fmt.Sprintf(
		" %s",
		boldgreen(trimPath(cwd, home)),
	))
}

func addGitSeg(segments []string) []string {
	gitDir := getGitDir()
	if len(gitDir) == 0 {
		return segments
	}
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return segments
	}
	ref, err := repo.Head()
	if err != nil {
		return segments
	}
	branch := strings.TrimPrefix(string(ref.Name()), "refs/heads/")
	return append(segments, fmt.Sprintf(
		" %s",
		boldwhite(branch),
		))
}

func makePrompt() string {
	cwd, _ = os.Getwd()
	home = os.Getenv("HOME")
	segments := addDirSeg([]string{})
	segments = addGitSeg(segments)

	return fmt.Sprintf(
		"\n%s %s",
		strings.Join(segments, " | "),
		promptSym,
	)
}

func main() {
	fmt.Println(makePrompt())
}
