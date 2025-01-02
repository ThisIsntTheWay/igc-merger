package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/twpayne/go-igc"
)

// Verify that merge candidates are permissible for merge. Non-zero exit on failure.
func ensureMergeSafety(igcs []igc.IGC) {
	for tlc, thisHRecord := range igcs[0].HRecordsByTLC {
		if igcs[1].HRecordsByTLC[tlc].Value != thisHRecord.Value {
			slog.Error("MERGE",
				"action", "verifyMergeCandidate",
				"error", "Header value mismatch",
				"header", tlc,
				"candidate1", thisHRecord.Value,
				"candidate2", igcs[1].HRecordsByTLC[tlc].Value,
			)
			os.Exit(1)
		}
	}
}

// Separate B records from the rest for a given io.Reader.
// 1st []string: B records, 2nd []string: other records, int: index of last non-B record
func isolateRecords(f io.Reader) ([]string, []string, int) {
	var bRecords, otherRecords []string
	var indexOfLastNonBRecord int

	foundLastNonBRecord := false
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "B") {
			foundLastNonBRecord = true
			bRecords = append(bRecords, s.Text())
		} else {
			// Ignore specific records as they'll be overwritten anyway
			skipThisLine := false
			for _, p := range []string{"A", "G"} {
				if strings.HasPrefix(s.Text(), p) {
					skipThisLine = true
					break
				}
			}
			if !foundLastNonBRecord {
				indexOfLastNonBRecord += 1
			}
			if !skipThisLine {
				otherRecords = append(otherRecords, s.Text())
			}
		}
	}
	if err := s.Err(); err != nil && err != io.EOF {
		panic(fmt.Errorf("error reading lines: %v", err))
	}

	return bRecords, otherRecords, indexOfLastNonBRecord
}

// Sort B records chronologically (oldest first)
func sortBRecords(recordsArray []string) {
	type bRecord struct {
		Timestamp int
		Record    string
	}

	var bRecords []bRecord
	for _, record := range recordsArray {
		timestamp, err := strconv.Atoi(record[1:7])
		if err != nil {
			slog.Error("MERGE", "action", "parseTimestamp", "error", err, "record", record)
			os.Exit(1)
		}

		bRecords = append(bRecords, bRecord{
			Timestamp: timestamp,
			Record:    record,
		})
	}

	sort.Slice(bRecords, func(i, j int) bool {
		return bRecords[i].Timestamp < bRecords[j].Timestamp
	})

	for i, bRecord := range bRecords {
		recordsArray[i] = bRecord.Record
	}
}

// Generates a custom A record
func generateARecord() string {
	return "AXXX MERGED IGC"
}

// Merge 2 IGC files (length of []candidates is capped at 2)
func mergeIGCs(candidates []string) []string {
	// Consume IGCs
	var readers []io.Reader
	var igcs []igc.IGC
	for _, candidate := range candidates[0:2] {
		f, err := os.Open(candidate)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// Buffering required as igc.Parse would consume f's io.Stream
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, f); err != nil {
			panic(err)
		}

		v, err := igc.Parse(bytes.NewReader(buf.Bytes()))
		if err != nil {
			slog.Error("MERGE", "action", "parseIgc", "error", err)
			os.Exit(1)
		}
		igcs = append(igcs, *v)

		readers = append(readers, bytes.NewReader(buf.Bytes()))
	}

	ensureMergeSafety(igcs)

	// Isolate B records from the rest.
	// We'll only retain non-B records of the first IGC
	var bRecords, otherRecords []string
	var indexOfLastNonBRecord int
	for i, c := range readers {
		b, o, ilr := isolateRecords(c)
		if i == 0 {
			indexOfLastNonBRecord = ilr
			otherRecords = append(otherRecords, o...)
		}

		bRecords = append(bRecords, b...)
	}
	sortBRecords(bRecords)

	// Reassemble
	var reassembledIgc []string
	reassembledIgc = append(reassembledIgc, generateARecord())
	reassembledIgc = append(reassembledIgc, otherRecords[0:indexOfLastNonBRecord-1]...)
	reassembledIgc = append(reassembledIgc, bRecords...)
	reassembledIgc = append(reassembledIgc, otherRecords[indexOfLastNonBRecord-1:]...)
	reassembledIgc = append(reassembledIgc, calculateChecksum(
		[]byte(strings.Join(reassembledIgc, "")),
	))

	return reassembledIgc
}
