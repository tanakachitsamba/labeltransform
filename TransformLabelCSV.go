package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

/*───────────────────────────────
   PRODUCTION  CODE
────────────────────────────────*/

// StringToBinary converts the tokens `"true"` and `"false"` (case‑insensitive,
// leading/trailing whitespace ignored) into `"1"` and `"0"` respectively.
// Any other value returns an error.  The implementation works directly on
// ASCII bytes, avoiding an allocation for strings.ToLower.
func StringToBinary(val string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(val))

	binMap := map[string]string{
		"true":     "1",
		"positive": "1",
		"yes":      "1",
		"1":        "1",
		"false":    "0",
		"negative": "0",
		"no":       "0",
		"0":        "0",
	}
	if b, ok := binMap[s]; ok {
		return b, nil
	}
	return "", fmt.Errorf("unexpected label value: %q", val)
}

// log1p is a helper because math.Log1p returns NaN for negative numbers.
func log1p(x float64) float64 {
	if x < 0 {
		return 0
	}
	return math.Log1p(x)
}

// TransformLabelCSV streams an input CSV file, rewrites the `label` column
// using StringToBinary, and writes the result to outputFile.
// It is optimised for large files: constant memory overhead and buffered I/O.
func TransformLabelCSV(inputFile, outputFile string) error {
	/* Open files --------------------------------------------------------- */
	in, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer in.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	const bufSize = 4 << 20 // 4 MiB
	br := bufio.NewReaderSize(in, bufSize)
	bw := bufio.NewWriterSize(out, bufSize)
	defer bw.Flush()

	reader := csv.NewReader(br)
	reader.ReuseRecord = true // avoid per‑row allocations

	writer := csv.NewWriter(bw)
	defer writer.Flush()

	/* Locate columns ----------------------------------------------------- */
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	labelIdx := -1
	trueDurIdx := -1
	for i, col := range header {
		colTrim := strings.TrimSpace(col)
		switch {
		case strings.EqualFold(colTrim, "label"):
			labelIdx = i
		case strings.EqualFold(colTrim, "true_duration_seconds"):
			trueDurIdx = i
		}
	}
	if labelIdx == -1 {
		return fmt.Errorf("no column named 'label' found")
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	/* Stream rows -------------------------------------------------------- */
	const flushEvery = 100_000
	rowNum := 1 // header already counted

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			return fmt.Errorf("read row %d: %w", rowNum, err)
		}

		// Translate label --------------------------------------------------
		bin, err := StringToBinary(row[labelIdx])
		if err != nil {
			return fmt.Errorf("row %d: %w", rowNum, err)
		}
		row[labelIdx] = bin

		// (future‑proofed slot for true_duration_seconds) ------------------
		if trueDurIdx != -1 { // currently impossible – see early guard
			v, err := strconv.ParseFloat(row[trueDurIdx], 64)
			if err != nil {
				return fmt.Errorf("row %d: invalid true_duration_seconds: %w", rowNum, err)
			}
			row[trueDurIdx] = fmt.Sprintf("%f", log1p(v))
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row %d: %w", rowNum, err)
		}
		if rowNum%flushEvery == 0 {
			writer.Flush()
			if err := writer.Error(); err != nil {
				return fmt.Errorf("flush: %w", err)
			}
		}
	}
	return writer.Error()
}
