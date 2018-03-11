package main

import (
	"fmt"
	"strings"
	"regexp"
	"io/ioutil"
	"log"
	"path/filepath"
	"flag"
)

func flatten(strings [][]string) []string {
	result := []string{}

	for _, text := range strings {
		result = append(result, text...)
	}

	return result
}

func collect(source string, format string) []string {
	r := regexp.MustCompile(format)
	found := r.FindAllStringSubmatch(source, -1)

	return flatten(found)
}

func dirwalk(dir string) []string {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        panic(err)
    }

    var paths []string
    for _, file := range files {
        if file.IsDir() {
            paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
            continue
        }
        paths = append(paths, filepath.Join(dir, file.Name()))
    }

    return paths
}

func filterString(strings []string, isIncluded func(string) bool) []string {
	var results []string
	for _, s := range strings {
		if isIncluded(s) {
			results = append(results, s)
		}
	}
	return results
}

func content(filename string) string {
	buffer, e := ioutil.ReadFile(filename)

	if e != nil {
		log.Fatal(e)
	}

	return string(buffer)
}

func pathExtension(filename string) string {
	components := strings.Split(filename, "/")
	return components[len(components) - 1]
}

func parseSwiftSource(filenames []string, format string) []string {
	var keywords []string

	for _, filename := range filenames {
		content := content(filename)	
		result := collect(content, format)

		for _, r := range result {
			keywords = append(keywords, pathExtension(filename) + "\t" + r)
		}
	}

	return keywords
}

func parseInterfaceBuilder(filenames []string) []string {
	var keywords []string

	for _, filename := range filenames {
		content := content(filename)	
		result := collect(content, `text=\".+?\"`)
		result = append(result, collect(content, `title=\".+?\"`)...)
		result = append(result, collect(content, `placeholder=\".+?\"`)...)

		for _, r := range result {
			keywords = append(keywords, pathExtension(filename) + "\t" + r)
		}
	}

	return keywords
}


func main() {
    flag.Parse()
    args := flag.Args()
	
	files := dirwalk(args[0])

	filterFile := func(filename string) func(string) bool {
		return func(s string) bool {
			return strings.Contains(s, filename)
		}
	}

	swiftSourcesNames := filterString(files, filterFile(".swift"))
	storyboardNames := filterString(files, filterFile(".storyboard"))
	xibNames := filterString(files, filterFile(".xib"))

	keywords := parseSwiftSource(swiftSourcesNames, `\".+?\"`)
	keywords = append(keywords, parseInterfaceBuilder(storyboardNames)...)
	keywords = append(keywords, parseInterfaceBuilder(xibNames)...)

	fmt.Println(strings.Join(keywords, "\n"))
}
