package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/twpayne/go-igc/civlovs"
)

var cl *civlovs.Client = civlovs.NewClient()

// Verify the integrity of an IGC file
func VerifyIgc(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("opening '%s' failed: %v", filePath, err)
	}

	_, r, err := cl.ValidateIGC(
		context.Background(),
		filePath,
		file,
	)
	if err != nil {
		return false, fmt.Errorf("validating '%s' failed: %v", filePath, err)
	}

	slog.Debug("VERIFY", "filePath", filePath, "serverResponse", r)
	return r.Passed(), nil
}
