package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/functions"
	"github.com/pivotal-golang/lager/chug"
)

var funcs map[string]func(Entries, string) error

var outputDir string

func init() {
	funcs = map[string]func(Entries, string) error{
		"fezzik-tasks":            functions.FezzikTasks,
		"vizzini-parallel-garden": functions.VizziniParallelGarden,
	}

	flag.StringVar(&outputDir, "output-dir", ".", "Output Directory to store plots")
	flag.Parse()
}

func main() {
	if len(flag.Args()) != 2 {
		PrintUsage()
		os.Exit(1)
	}

	f, ok := funcs[flag.Args()[0]]

	if !ok {
		PrintUsage()
		os.Exit(1)
	}

	entries, err := LoadEntries(flag.Args()[1])
	if err != nil {
		fmt.Printf("Failed to load %s\n", flag.Args()[1])
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = f(entries, outputDir)
	if err != nil {
		fmt.Printf("Function %s failed", flag.Args()[0])
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func PrintUsage() {
	fmt.Println("log_measure FUNCTION LOG_FILE")
	fmt.Println("-----------------------------")
	fmt.Println("Available functions:")
	functionNames := []string{}
	for k := range funcs {
		functionNames = append(functionNames, k)
	}
	sort.Strings(functionNames)
	for _, name := range functionNames {
		fmt.Println("\t" + name)
	}
}

func LoadEntries(filename string) (Entries, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	out := make(chan chug.Entry)
	go chug.Chug(file, out)

	entries := Entries{}
	for chugEntry := range out {
		entry, err := NewEntry(chugEntry)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
