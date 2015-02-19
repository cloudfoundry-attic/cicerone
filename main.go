package main

import (
	"fmt"
	"os"
	"sort"

	. "github.com/onsi/sommelier/dsl"
	"github.com/onsi/sommelier/functions"
	"github.com/pivotal-golang/lager/chug"
)

var funcs map[string]func(Entries) error

func init() {
	funcs = map[string]func(Entries) error{
		"playground": functions.Playground,
	}
}

func main() {

	if len(os.Args) != 3 {
		PrintUsage()
		os.Exit(1)
	}

	f, ok := funcs[os.Args[1]]

	if !ok {
		PrintUsage()
		os.Exit(1)
	}

	entries, err := LoadEntries(os.Args[2])
	if err != nil {
		fmt.Printf("Failed to load %s\n", os.Args[2])
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = f(entries)
	if err != nil {
		fmt.Printf("Function %s failed", os.Args[1])
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
