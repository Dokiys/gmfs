package conv

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenTpyConv(t *testing.T) {
	const X, Y = "X", "Y"
	const wantFile = "want"

	tests := []struct {
		name string
	}{
		{"Basic"},
		{"Nested"},
		{"PkgStruct_basic"},
		{"Pointer_X"},
		{"Pointer_XY"},
		{"Pointer_Y"},
		// TODO[Dokiy] 2022/9/30: arr, slice
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gopath, err := filepath.Abs("testdata/tpy_conv/")
			if err != nil {
				t.Fatalf("%s: get gopath err:%s", tt.name, err)
			}
			cfg := &packages.Config{
				Context: context.Background(),
				Mode:    packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
				Env:     append(os.Environ(), "GOPATH="+gopath),
			}

			pkgs, err := packages.Load(cfg, filepath.Join(gopath, tt.name))
			if err != nil {
				t.Fatalf("%s: incorrect test src: %s", tt.name, err)
			}
			for i, pkg := range pkgs {
				x := pkg.Types.Scope().Lookup(X)
				y := pkg.Types.Scope().Lookup(Y)
				if x == nil || y == nil {
					continue
				}

				tcg := &TypConvGen{
					g:        newGener(""),
					pkgAlias: parseImportAlias(pkg.Syntax[i]),
					ignore:   nil,
				}
				tcg.Gen(NewTypCtx(x.Name(), y.Name(), x.Type(), y.Type()))

				got, err := format.Source([]byte(tcg.g.string()))
				if err != nil {
					t.Logf("got:\n %s\n", tcg.g.string())
					t.Fatalf("%s: format genSrc err: %s", wantFile, err)
				}

				expected, _ := os.ReadFile(filepath.Join(gopath, tt.name, wantFile))
				if strings.Compare(string(got), string(expected)) != 0 {
					t.Errorf("result did not meet expections:\n %s\n", diff(string(expected), string(got)))
				}
				return
			}
			t.Fatalf("test must declare both %s and %s", X, Y)
		})
	}
}

// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Chunk represents a piece of the diff.  A chunk will not have both added and
// deleted lines.  Equal lines are always after any added or deleted lines.
// A Chunk may or may not have any lines in it, especially for the first or last
// chunk in a computation.
type Chunk struct {
	Added   []string
	Deleted []string
	Equal   []string
}

func (c *Chunk) empty() bool {
	return len(c.Added) == 0 && len(c.Deleted) == 0 && len(c.Equal) == 0
}

// diff returns a string containing a line-by-line unified diff of the linewise
// changes required to make A into B.  Each line is prefixed with '+', '-', or
// ' ' to indicate if it should be added, removed, or is correct respectively.
func diff(A, B string) string {
	aLines := strings.Split(A, "\n")
	bLines := strings.Split(B, "\n")

	chunks := diffChunks(aLines, bLines)

	buf := new(bytes.Buffer)
	for _, c := range chunks {
		for _, line := range c.Added {
			fmt.Fprintf(buf, "+%s\n", line)
		}
		for _, line := range c.Deleted {
			fmt.Fprintf(buf, "-%s\n", line)
		}
		for _, line := range c.Equal {
			fmt.Fprintf(buf, " %s\n", line)
		}
	}
	return strings.TrimRight(buf.String(), "\n")
}

// diffChunks uses an O(D(N+M)) shortest-edit-script algorithm
// to compute the edits required from A to B and returns the
// edit chunks.
func diffChunks(a, b []string) []Chunk {
	// algorithm: http://www.xmailserver.org/diff2.pdf

	// We'll need these quantities a lot.
	alen, blen := len(a), len(b) // M, N

	// At most, it will require len(a) deletions and len(b) additions
	// to transform a into b.
	maxPath := alen + blen // MAX
	if maxPath == 0 {
		// degenerate case: two empty lists are the same
		return nil
	}

	// Store the endpoint of the path for diagonals.
	// We store only the a index, because the b index on any diagonal
	// (which we know during the loop below) is aidx-diag.
	// endpoint[maxPath] represents the 0 diagonal.
	//
	// Stated differently:
	// endpoint[d] contains the aidx of a furthest reaching path in diagonal d
	endpoint := make([]int, 2*maxPath+1) // V

	saved := make([][]int, 0, 8) // Vs
	save := func() {
		dup := make([]int, len(endpoint))
		copy(dup, endpoint)
		saved = append(saved, dup)
	}

	var editDistance int // D
dLoop:
	for editDistance = 0; editDistance <= maxPath; editDistance++ {
		// The 0 diag(onal) represents equality of a and b.  Each diagonal to
		// the left is numbered one lower, to the right is one higher, from
		// -alen to +blen.  Negative diagonals favor differences from a,
		// positive diagonals favor differences from b.  The edit distance to a
		// diagonal d cannot be shorter than d itself.
		//
		// The iterations of this loop cover either odds or evens, but not both,
		// If odd indices are inputs, even indices are outputs and vice versa.
		for diag := -editDistance; diag <= editDistance; diag += 2 { // k
			var aidx int // x
			switch {
			case diag == -editDistance:
				// This is a new diagonal; copy from previous iter
				aidx = endpoint[maxPath-editDistance+1] + 0
			case diag == editDistance:
				// This is a new diagonal; copy from previous iter
				aidx = endpoint[maxPath+editDistance-1] + 1
			case endpoint[maxPath+diag+1] > endpoint[maxPath+diag-1]:
				// diagonal d+1 was farther along, so use that
				aidx = endpoint[maxPath+diag+1] + 0
			default:
				// diagonal d-1 was farther (or the same), so use that
				aidx = endpoint[maxPath+diag-1] + 1
			}
			// On diagonal d, we can compute bidx from aidx.
			bidx := aidx - diag // y
			// See how far we can go on this diagonal before we find a difference.
			for aidx < alen && bidx < blen && a[aidx] == b[bidx] {
				aidx++
				bidx++
			}
			// Store the end of the current edit chain.
			endpoint[maxPath+diag] = aidx
			// If we've found the end of both inputs, we're done!
			if aidx >= alen && bidx >= blen {
				save() // save the final path
				break dLoop
			}
		}
		save() // save the current path
	}
	if editDistance == 0 {
		return nil
	}
	chunks := make([]Chunk, editDistance+1)

	x, y := alen, blen
	for d := editDistance; d > 0; d-- {
		endpoint := saved[d]
		diag := x - y
		insert := diag == -d || (diag != d && endpoint[maxPath+diag-1] < endpoint[maxPath+diag+1])

		x1 := endpoint[maxPath+diag]
		var x0, xM, kk int
		if insert {
			kk = diag + 1
			x0 = endpoint[maxPath+kk]
			xM = x0
		} else {
			kk = diag - 1
			x0 = endpoint[maxPath+kk]
			xM = x0 + 1
		}
		y0 := x0 - kk

		var c Chunk
		if insert {
			c.Added = b[y0:][:1]
		} else {
			c.Deleted = a[x0:][:1]
		}
		if xM < x1 {
			c.Equal = a[xM:][:x1-xM]
		}

		x, y = x0, y0
		chunks[d] = c
	}
	if x > 0 {
		chunks[0].Equal = a[:x]
	}
	if chunks[0].empty() {
		chunks = chunks[1:]
	}
	if len(chunks) == 0 {
		return nil
	}
	return chunks
}
