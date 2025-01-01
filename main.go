package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var outputFile string = "./merged.igc"

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Must provide 2 arguments: app <file1.igc> <file2.igc>")
		os.Exit(1)
	}

	// File validations
	args := os.Args[1:]
	for i, arg := range args {
		if !strings.HasSuffix(strings.ToLower(arg), "igc") {
			panic(fmt.Sprintf("Argument %d (%s) must be an IGC file", i+1, arg))
		}
	}

	slog.Info("MAIN", "action", "verifyIgcFiles", "targets", []string{args[0], args[1]})
	for _, a := range args {
		valid, e := VerifyIgc(a)
		if e != nil {
			os.Exit(1)
		} else if !valid {
			slog.Warn("MAIN", "file", a, "verified", valid)
		} else {
			slog.Info("MAIN", "file", a, "verified", valid)
		}
	}

	mergedIGCs := mergeIGCs(args)
	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("MAIN",
			"action", "writeIgc",
			"filePath", outputFile,
			"error", err,
		)
		fmt.Println(mergedIGCs)
		os.Exit(1)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range mergedIGCs {
		if _, e := w.WriteString(l + "\n"); e != nil {
			panic(e)
		}
	}
	w.Flush()

	slog.Info("MAIN",
		"action", "writeIgc",
		"filePath", outputFile,
		"success", true,
	)
}
