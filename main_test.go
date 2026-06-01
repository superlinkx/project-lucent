// Copyright 2026 Alyx Holms
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

package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user/lucent/internal/analyzer"
)

func TestRunAnalysisReport(t *testing.T) {
	// Construct a call graph:
	// entry1 -> childA (complexity 2)
	// entry2 -> childA (complexity 2), childB (complexity 3)
	// childA -> leaf (complexity 4)
	
	graph := map[string]*analyzer.FunctionNode{
		"main.entry1": {
			ID:         "main.entry1",
			Package:    "main",
			Name:       "entry1",
			Complexity: 1,
			IsEntry:    true,
			Calls:      []string{"main.childA"},
		},
		"main.entry2": {
			ID:         "main.entry2",
			Package:    "main",
			Name:       "entry2",
			Complexity: 1,
			IsEntry:    true,
			Calls:      []string{"main.childA", "main.childB"},
		},
		"main.childA": {
			ID:         "main.childA",
			Package:    "main",
			Name:       "childA",
			Complexity: 2,
			Calls:      []string{"main.leaf"},
		},
		"main.childB": {
			ID:         "main.childB",
			Package:    "main",
			Name:       "childB",
			Complexity: 3,
			Calls:      []string{},
		},
		"main.leaf": {
			ID:         "main.leaf",
			Package:    "main",
			Name:       "leaf",
			Complexity: 4,
			Calls:      []string{},
		},
	}

	entryPoints := []string{"entry1", "entry2"}
	report := RunAnalysisReport(graph, entryPoints, nil)

	// Validate Rolled Complexity rankings under Ungrouped:
	// entry1: entry1 (1) + childA (2) + leaf (4) = 7. Reachable nodes: 3.
	// entry2: entry2 (1) + childA (2) + childB (3) + leaf (4) = 10. Reachable nodes: 4.
	// Sorted: entry1 (7) then entry2 (10)
	assert.Contains(t, report, "ROUTE GROUP: UNGROUPED")
	assert.Contains(t, report, "1. main.entry1")
	assert.Contains(t, report, "Combined Rolled Complexity: 7")
	assert.Contains(t, report, "2. main.entry2")
	assert.Contains(t, report, "Combined Rolled Complexity: 10")

	// Validate Shared Dependencies:
	// childA and leaf are reachable from both entry1 and entry2.
	assert.Contains(t, report, "Shared by 2 endpoints: [main.entry1, main.entry2]")
	assert.Contains(t, report, "main.childA")
	assert.Contains(t, report, "main.leaf")
}

func TestWriteOutput(t *testing.T) {
	// Test stdout path (empty filepath)
	writeOutput([]byte("stdout test"), "")

	// Test file path
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")
	writeOutput([]byte("file test"), filePath)

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, "file test", string(data))
}

func TestMainFunc(t *testing.T) {
	// Setup a temporary workspace
	tempDir := t.TempDir()
	goModContent := "module github.com/user/tempmod\ngo 1.20\n"
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	mainGoContent := "package main\nfunc main() {}\n"
	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGoContent), 0644)

	// Change working directory to tempDir to prevent overwriting/deleting files in the user workspace
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Mock command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset global flag state
	oldCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldCommandLine }()

	formats := []string{"analysis", "json", "c4", "html", "all"}
	for _, format := range formats {
		os.Args = []string{"lucent", "-dir", ".", "-entry", "main", "-format", format, "-out", ""}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		
		// Run main() - should execute scanning and complete successfully
		main()
	}
}
