package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

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
	red        = color("\001\033[31m\002%s\001\033[0m\002")
	green      = color("\001\033[32m\002%s\001\033[0m\002")
	boldgreen  = color("\001\033[1;32m\002%s\001\033[0m\002")
	boldcyan   = color("\001\033[1;36m\002%s\001\033[0m\002")
	boldwhite  = color("\001\033[1;37m\002%s\001\033[0m\002")
	boldpurple = color("\001\033[1;35m\002%s\001\033[0m\002")
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
	cwd  string
	home string
)

func trimPath(cwd, home string) string {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}

	path := strings.Replace(cwd, home, "~", 1)
	if len(path) <= width/4 {
		return path
	}
	items := strings.Split(path, "/")
	truncItems := []string{}
	for i, item := range items {
		if i != (len(items)-1) && i != 0 {
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

func addUserSeg(segments []string) []string {
	currentUser, err := user.Current()
	if err != nil {
		return segments
	}

	shortHost := regexp.MustCompile("\\..*$")
	hostname, err := os.Hostname()
	if err != nil {
		return segments
	}

	return append(segments, fmt.Sprintf(
		green(" %s@%s"),
		currentUser.Username,
		shortHost.ReplaceAllString(hostname, ""),
	))
}

func addDirSeg(segments []string) []string {
	return append(segments, fmt.Sprintf(
		boldgreen(" %s"),
		trimPath(cwd, home),
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
		boldwhite(" %s"),
		branch,
	))
}

func trimSegments(segments []string) []string {
	trimRe := regexp.MustCompile("\001[^\002]*\002")
	width, _, err := term.GetSize(0)
	if err != nil {
		return segments
	}
	for {
		promptlen := utf8.RuneCountInString(trimRe.ReplaceAllString(strings.Join(segments, " | "), ""))
		if len(segments) == 0 || promptlen < width/2 {
			return segments
		}
		segments = segments[1:]
	}
}

func makePrompt() string {
	cwd, _ = os.Getwd()
	home = os.Getenv("HOME")
	segments := addUserSeg([]string{})
	segments = addDirSeg(segments)
	segments = addGitSeg(segments)
	segments = trimSegments(segments)

	return fmt.Sprintf(
		"\n%s %s\001 \002",
		strings.Join(segments, " | "),
		promptSym,
	)
}

func main() {
	fmt.Print(makePrompt())
}
