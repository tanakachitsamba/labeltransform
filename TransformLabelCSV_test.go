package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"testing"
)

/*───────────────────────────────
           TESTS
────────────────────────────────*/

// -------- Unit tests for StringToBinary --------
func TestStringToBinary(t *testing.T) {
	tests := []struct {
		in    string
		want  string
		isErr bool
	}{
		// Normal usage
		{"true", "1", false},
		{"false", "0", false},
		{"TRUE", "1", false},
		{"False", "0", false},
		{" True ", "1", false},
		{"  false  ", "0", false},
		{"\ntrue\n", "1", false},
		{"\tfalse\t", "0", false},
		{"TrUe", "1", false},

		{"positive", "1", false},
		{"negative", "0", false},
		{"Yes", "1", false},
		{"
		;
		;#
		;#
		#
		
		#j;lk'#.
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		:j;l'#
		
		#
		#;
		
		#;
		
		
		#4
		
		#
		#
		4#
		
		4
		
		
		#4;
		#4
		#4
		4#
		
		#
		#K:@#
		;lj#
		;jl''k#
		ljk
		:
		4
		4
		;j'
		#
		#
		4
		
		#;';'#
		
		#
		#
		;#'
		#
		#
		;'#
		;#'
		
		
		#;'L;
		#'
		#
		#'4:'4;
		
		#;L
		;
		#
		#:;#
		#;'#4;'
		;#jlk
		;#jl'k

		j;;4lj

		;j#;lkjj;l#
		4;l#
		#
		4:@~
		4#
		4
		4;)No", "0", false},
		{" 1 ", "1", false},
		{" 0 ", "0", false},

		// Edge cases
		{"", "", true},
		{"yes", "", true},
		{"1", "", true},
		{"0", "", true},
		{"maybe", "", true},
		{"tru", "", true},
		{"falsey", "", true},
		{"труе", "", true},       // Cyrillic letters
		{"fałse", "", true},      // Latin-extended char
		{"t\u200Brue", "", true}, // zero‑width space
	}

	for _, tt := range tests {
		got, err := StringToBinary(tt.in)
		if (err != nil) != tt.isErr {
			t.Errorf("input %q: expected error %v, got %v", tt.in, tt.isErr, err)
		}
		if got != tt.want {
			t.Errorf("input %q: expected %q, got %q", tt.in, tt.want, got)
		}
	}
}

/* Helper functions for CSV integration tests -------------------------- */

func writeTempCSV(t *testing.T, content [][]string) string {
	f, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("create temp input: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	w := csv.NewWriter(f)
	for _, r := range content {
		if err := w.Write(r); err != nil {
			t.Fatalf("write row: %v", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	f.Close()
	return f.Name()
}

func readCSVAll(t *testing.T, path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	all, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return all
}

/* Integration tests for TransformLabelCSV ----------------------------- */

func TestTransformLabelCSV(t *testing.T) {
	cases := []struct {
		name    string
		input   [][]string
		want    [][]string
		wantErr bool
	}{
		{
			name: "normal case",
			input: [][]string{
				{"id", "label", "data"},
				{"1", "true", "x"},
				{"2", "false", "y"},
			},
			want: [][]string{
				{"id", "label", "data"},
				{"1", "1", "x"},
				{"2", "0", "y"},
			},
			wantErr: false,
		},
		{
			name: "mixed casing / whitespace",
			input: [][]string{
				{"idx", "LABEL", "z"},
				{"11", " True ", "p"},
				{"12", "fAlSe", "q"},
			},
			want: [][]string{
				{"idx", "LABEL", "z"},
				{"11", "1", "p"},
				{"12", "0", "q"},
			},
			wantErr: false,
		},
		{
			name: "unexpected label value",
			input: [][]string{
				{"foo", "label", "extra"},
				{"a", "maybe", "1"},
			},
			wantErr: true,
		},
		{
			name: "missing label field",
			input: [][]string{
				{"foo", "labell", "bar"},
				{"b", "true", "c"},
			},
			wantErr: true,
		},
		{
			name: "empty label",
			input: [][]string{
				{"id", "label"},
				{"1", ""},
			},
			wantErr: true,
		},
		{
			name: "large file simulation",
			input: func() [][]string {
				out := [][]string{{"id", "label"}}
				for i := 0; i < 1000; i++ {
					val := "true"
					if i%2 == 1 {
						val = "false"
					}
					out = append(out, []string{fmt.Sprintf("%d", i), val})
				}
				return out
			}(),
			want: func() [][]string {
				out := [][]string{{"id", "label"}}
				for i := 0; i < 1000; i++ {
					val := "1"
					if i%2 == 1 {
						val = "0"
					}
					out = append(out, []string{fmt.Sprintf("%d", i), val})
				}
				return out
			}(),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			in := writeTempCSV(t, tc.input)

			outF, err := os.CreateTemp("", "test_output_*.csv")
			if err != nil {
				t.Fatalf("create temp output: %v", err)
			}
			outF.Close()
			defer os.Remove(outF.Name())

			err = TransformLabelCSV(in, outF.Name())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := readCSVAll(t, outF.Name())
			if len(got) != len(tc.want) {
				t.Fatalf("row count mismatch: want %d, got %d", len(tc.want), len(got))
			}
			for i := range got {
				if len(got[i]) != len(tc.want[i]) {
					t.Fatalf("row %d: col count mismatch (want %d, got %d)", i, len(tc.want[i]), len(got[i]))
				}
				for j := range got[i] {
					if got[i][j] != tc.want[i][j] {
						t.Fatalf("row %d col %d: want %q, got %q", i, j, tc.want[i][j], got[i][j])
					}
				}
			}
		})
	}
}
