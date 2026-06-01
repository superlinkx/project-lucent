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

package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superlinkx/project-lucent/internal/analyzer"
)

func TestModuleAnalyzerIntegration(t *testing.T) {
	// 1. Set up a temporary Go workspace
	tempDir := t.TempDir()

	// Write go.mod
	goModContent := `module github.com/user/tempmod

go 1.20
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	// Write main.go (root package)
	mainGoContent := `package main

import (
	"fmt"
	"github.com/user/tempmod/pkg1"
)

func main() {
	pkg1.Hello()
	localHelper()
}

func localHelper() {
	if true {
		fmt.Println("Complexity point")
	}
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGoContent), 0644)
	require.NoError(t, err)

	// Write pkg1/hello.go (sub-package)
	err = os.MkdirAll(filepath.Join(tempDir, "pkg1"), 0755)
	require.NoError(t, err)

	pkg1GoContent := `package pkg1

import "github.com/user/tempmod/pkg2"

type Greeter struct{}

func Hello() {
	pkg2.DoWork()
}

func (g Greeter) Greet() {
	// Monospace method declaration to test receiver ID builder
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg1", "hello.go"), []byte(pkg1GoContent), 0644)
	require.NoError(t, err)

	// Write pkg2/work.go (sub-package with conditional/loops complexity)
	err = os.MkdirAll(filepath.Join(tempDir, "pkg2"), 0755)
	require.NoError(t, err)

	pkg2GoContent := `package pkg2

func DoWork() {
	for i := 0; i < 5; i++ {
		switch i {
		case 1:
			println("one")
		case 2:
			println("two")
		default:
			println("other")
		}
	}
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg2", "work.go"), []byte(pkg2GoContent), 0644)
	require.NoError(t, err)

	// 2. Initialize ModuleAnalyzer
	ma, err := analyzer.NewModuleAnalyzer(tempDir)
	require.NoError(t, err)
	
	// Scan the workspace
	err = ma.Scan()
	require.NoError(t, err)

	// Verify method registration exists in scanned functions
	methodID := "github.com/user/tempmod/pkg1.(Greeter).Greet"
	assert.Contains(t, ma.Funcs, methodID, "Receiver methods should be scanned and registered successfully")

	// 4. Analyze the call graph starting from main.main
	graph := ma.Analyze([]string{"main"})
	require.NotEmpty(t, graph)

	// 5. Verify the extracted call graph nodes and properties
	
	// main.main node
	mainNode, ok := graph["github.com/user/tempmod.main"]
	require.True(t, ok, "main.main function should be present in the call graph")
	assert.True(t, mainNode.IsEntry)
	assert.Equal(t, 1, mainNode.Complexity)
	assert.Contains(t, mainNode.Calls, "github.com/user/tempmod.localHelper")
	assert.Contains(t, mainNode.Calls, "github.com/user/tempmod/pkg1.Hello")

	// main.localHelper node
	helperNode, ok := graph["github.com/user/tempmod.localHelper"]
	require.True(t, ok, "localHelper should be present")
	assert.Equal(t, 2, helperNode.Complexity, "Complexity should be 2 due to if statement")
	assert.Contains(t, helperNode.ExternalCalls, "fmt.Println")

	// pkg1.Hello node
	helloNode, ok := graph["github.com/user/tempmod/pkg1.Hello"]
	require.True(t, ok, "pkg1.Hello should be present")
	assert.Contains(t, helloNode.Calls, "github.com/user/tempmod/pkg2.DoWork")

	// pkg2.DoWork node
	workNode, ok := graph["github.com/user/tempmod/pkg2.DoWork"]
	require.True(t, ok, "pkg2.DoWork should be present")
	// Complexity: 1 (base) + 1 (for) + 1 (case 1) + 1 (case 2) = 4
	assert.Equal(t, 4, workNode.Complexity, "Complexity should be 4 due to loop and cases")

	// 6. Test cyclic calls are handled gracefully without infinite loops
	cyclicGoContent := `package pkg2
import "github.com/user/tempmod/pkg1"
func DoWork() {
	pkg1.Hello()
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg2", "work.go"), []byte(cyclicGoContent), 0644)
	require.NoError(t, err)

	maCyclic, err := analyzer.NewModuleAnalyzer(tempDir)
	require.NoError(t, err)
	err = maCyclic.Scan()
	require.NoError(t, err)

	// This should not hang or panic
	cyclicGraph := maCyclic.Analyze([]string{"main"})
	assert.NotEmpty(t, cyclicGraph)
}

func TestModuleAnalyzerInterfaceTracing(t *testing.T) {
	tempDir := t.TempDir()

	// Write go.mod
	goModContent := `module github.com/user/tempmod

go 1.20
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	// Write pkg2/service.go
	err = os.MkdirAll(filepath.Join(tempDir, "pkg2"), 0755)
	require.NoError(t, err)

	pkg2Content := `package pkg2

type Database interface {
	GetUsers()
}

type pgDatabase struct{}

func (db pgDatabase) GetUsers() {
	// Concrete implementation
}

func NewpgDatabase() Database {
	return pgDatabase{}
}

func RunService(db Database) {
	db.GetUsers()
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg2", "service.go"), []byte(pkg2Content), 0644)
	require.NoError(t, err)

	// Write main.go
	mainContent := `package main

import "github.com/user/tempmod/pkg2"

func main() {
	db := pkg2.NewpgDatabase()
	pkg2.RunService(db)
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	require.NoError(t, err)

	// Initialize ModuleAnalyzer
	ma, err := analyzer.NewModuleAnalyzer(tempDir)
	require.NoError(t, err)

	err = ma.Scan()
	require.NoError(t, err)

	// Verify scanned functions contains the concrete method
	concreteMethodID := "github.com/user/tempmod/pkg2.(pgDatabase).GetUsers"
	assert.Contains(t, ma.Funcs, concreteMethodID)

	// Analyze call graph
	graph := ma.Analyze([]string{"main"})
	require.NotEmpty(t, graph)

	// Verify that RunService treats the interface call as an external boundary call
	runServiceNode, ok := graph["github.com/user/tempmod/pkg2.RunService"]
	require.True(t, ok, "RunService should be in the graph")
	assert.Contains(t, runServiceNode.BoundaryCalls, "db.GetUsers", "RunService should treat db.GetUsers as an external call boundary")
	assert.NotContains(t, runServiceNode.Calls, concreteMethodID, "RunService should not descend into concrete receiver method implementations")
}

