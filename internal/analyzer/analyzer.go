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

package analyzer

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Block represents a basic control-flow block in a function.
type Block struct {
	ID    string `json:"id"`
	Type  string `json:"type"`  // "entry", "stmt", "if_cond", "then", "else", "loop_header", "loop_body", "done"
	Label string `json:"label"` // Description of the statements
}

// CFGEdge represents a control-flow path between basic blocks.
type CFGEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"` // e.g., "true", "false", "next"
}

// CFG represents the Control Flow Graph of a single function.
type CFG struct {
	Blocks []Block   `json:"blocks"`
	Edges  []CFGEdge `json:"edges"`
}

// FunctionNode represents a function in the module call graph.
type FunctionNode struct {
	ID            string   `json:"id"`                    // Fully qualified function ID
	Package       string   `json:"package"`               // Importing package path
	Name          string   `json:"name"`                  // Function name
	Complexity    int      `json:"cyclomatic_complexity"` // Cyclomatic complexity score
	Calls         []string `json:"calls"`                 // IDs of internal functions called
	ExternalCalls []string `json:"external_calls"`        // IDs of external calls (std/third-party)
	BoundaryCalls []string `json:"boundary_calls"`        // Dynamic method calls (boundaries)
	CFG           CFG      `json:"cfg"`                   // Function's local Control Flow Graph
	IsEntry       bool     `json:"is_entry"`              // Is this a target entry-point?
}

// ModuleAnalyzer orchestrates the Go AST parsing and call graph extraction.
type ModuleAnalyzer struct {
	RootDir     string
	ModulePath  string
	Fileset     *token.FileSet
	Funcs       map[string]*FunctionNode
	astFuncs    map[string]*ast.FuncDecl
	funcImports map[string]map[string]string
}

// NewModuleAnalyzer initializes a analyzer with a root workspace path.
func NewModuleAnalyzer(rootDir string) (*ModuleAnalyzer, error) {
	absDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	modPath := discoverModulePath(absDir)
	return &ModuleAnalyzer{
		RootDir:     absDir,
		ModulePath:  modPath,
		Fileset:     token.NewFileSet(),
		Funcs:       make(map[string]*FunctionNode),
		astFuncs:    make(map[string]*ast.FuncDecl),
		funcImports: make(map[string]map[string]string),
	}, nil
}

// discoverModulePath extracts the Go module name from go.mod.
func discoverModulePath(dir string) string {
	goModPath := filepath.Join(dir, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "lucent" // Fallback name
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^module\s+(.+)$`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if match := re.FindStringSubmatch(line); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return "lucent"
}

// Scan codebase registers all declared functions and methods across the local module.
func (a *ModuleAnalyzer) Scan() error {
	err := filepath.Walk(a.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories, vendor, hidden files, and test files
		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		return a.parseGoFile(path)
	})
	return err
}

// parseGoFile parses a single Go source file and registers its function declarations.
func (a *ModuleAnalyzer) parseGoFile(path string) error {
	fileAST, err := parser.ParseFile(a.Fileset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Calculate import path relative to the module root
	relDir, err := filepath.Rel(a.RootDir, filepath.Dir(path))
	if err != nil {
		return err
	}

	pkgPath := a.ModulePath
	if relDir != "." {
		pkgPath = a.ModulePath + "/" + filepath.ToSlash(relDir)
	}

	// Map of package names/aliases to import paths inside this specific file
	imports := make(map[string]string)
	for _, imp := range fileAST.Imports {
		val := strings.Trim(imp.Path.Value, `"`)
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			parts := strings.Split(val, "/")
			name = parts[len(parts)-1]
		}
		imports[name] = val
	}

	for _, decl := range fileAST.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcID := a.buildFuncID(pkgPath, funcDecl)
		a.astFuncs[funcID] = funcDecl
		a.funcImports[funcID] = imports

		// Calculate complexity
		complexity := calculateComplexity(funcDecl)

		// Build a simplified CFG
		cfg := buildSyntacticCFG(funcDecl)

		a.Funcs[funcID] = &FunctionNode{
			ID:            funcID,
			Package:       pkgPath,
			Name:          funcDecl.Name.Name,
			Complexity:    complexity,
			CFG:           cfg,
			Calls:         []string{},
			ExternalCalls: []string{},
			BoundaryCalls: []string{},
		}
	}

	return nil
}

func (a *ModuleAnalyzer) buildFuncID(pkgPath string, decl *ast.FuncDecl) string {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return pkgPath + "." + decl.Name.Name
	}
	// It's a method call. Resolve receiver type name
	var recvType string
	switch t := decl.Recv.List[0].Type.(type) {
	case *ast.Ident:
		recvType = t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			recvType = id.Name
		}
	}
	if recvType != "" {
		return fmt.Sprintf("%s.(%s).%s", pkgPath, recvType, decl.Name.Name)
	}
	return pkgPath + "." + decl.Name.Name
}

// Analyze constructs the call graph starting from entry points.
func (a *ModuleAnalyzer) Analyze(entryPoints []string) map[string]*FunctionNode {
	resultGraph := make(map[string]*FunctionNode)
	visited := make(map[string]bool)
	var queue []string

	// Register entry points
	for _, ep := range entryPoints {
		// Try resolving partial match or direct ID
		resolvedEP := a.resolveFunctionID(ep)
		if resolvedEP != "" {
			if node, exists := a.Funcs[resolvedEP]; exists {
				node.IsEntry = true
				queue = append(queue, resolvedEP)
			}
		}
	}

	// BFS traversal
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]

		if visited[currID] {
			continue
		}
		visited[currID] = true

		node, exists := a.Funcs[currID]
		if !exists {
			continue
		}

		// Discover calls for this function
		astDecl := a.astFuncs[currID]
		if astDecl != nil {
			a.traceCalls(node, astDecl)
		}

		resultGraph[currID] = node

		// Queue internal children
		for _, childID := range node.Calls {
			if !visited[childID] {
				queue = append(queue, childID)
			}
		}
	}

	return resultGraph
}

func (a *ModuleAnalyzer) resolveFunctionID(name string) string {
	// Full match
	if _, exists := a.Funcs[name]; exists {
		return name
	}
	// Try partial matching by func name
	for id, node := range a.Funcs {
		if node.Name == name || strings.HasSuffix(id, "."+name) {
			return id
		}
	}
	return ""
}

func (a *ModuleAnalyzer) traceCalls(node *FunctionNode, decl *ast.FuncDecl) {
	if decl.Body == nil {
		return
	}

	// Statically inspect AST of body
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		switch fun := call.Fun.(type) {
		case *ast.Ident:
			// Package-level call
			callID := node.Package + "." + fun.Name
			if _, exists := a.Funcs[callID]; exists {
				node.Calls = append(node.Calls, callID)
			} else {
				node.ExternalCalls = append(node.ExternalCalls, fun.Name)
			}

		case *ast.SelectorExpr:
			// pkg.Func() or obj.Method()
			imports, hasImports := a.funcImports[node.ID]
			isPkgCall := false

			if xIdent, ok := fun.X.(*ast.Ident); ok && hasImports {
				pkgName := xIdent.Name
				targetFunc := fun.Sel.Name

				if importPath, isImported := imports[pkgName]; isImported {
					isPkgCall = true
					fullCallID := ""
					for localID, localNode := range a.Funcs {
						if localNode.Package == importPath && localNode.Name == targetFunc && !strings.Contains(localID, ".(") {
							fullCallID = localID
							break
						}
					}

					if fullCallID != "" {
						node.Calls = append(node.Calls, fullCallID)
					} else {
						node.ExternalCalls = append(node.ExternalCalls, fmt.Sprintf("%s.%s", pkgName, targetFunc))
					}
				}
			}

			if !isPkgCall {
				node.BoundaryCalls = append(node.BoundaryCalls, formatSelectorCall(fun))
			}
		}

		return true
	})

	node.Calls = uniqueStrings(node.Calls)
	node.ExternalCalls = uniqueStrings(node.ExternalCalls)
	node.BoundaryCalls = uniqueStrings(node.BoundaryCalls)
}

// calculateComplexity implements standard AST-based cyclomatic complexity.
func calculateComplexity(decl *ast.FuncDecl) int {
	if decl.Body == nil {
		return 0
	}

	complexity := 1
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			// Switch case. Exclude default case if possible.
			if len(node.List) > 0 { // Empty list represents default
				complexity++
			}
		case *ast.CommClause:
			// Select case. Exclude default case if possible.
			if node.Comm != nil { // Nil comm represents default
				complexity++
			}
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

// buildSyntacticCFG creates a simplified Control Flow Graph using Go AST constructs.
func buildSyntacticCFG(decl *ast.FuncDecl) CFG {
	var cfg CFG
	if decl.Body == nil {
		return cfg
	}

	blocks := []Block{
		{ID: "B_entry", Type: "entry", Label: fmt.Sprintf("Entry: %s", decl.Name.Name)},
	}
	edges := []CFGEdge{}

	currBlockID := "B_entry"
	stmtCount := 0
	blockCounter := 1

	var walkStatements func(stmts []ast.Stmt)
	walkStatements = func(stmts []ast.Stmt) {
		for _, stmt := range stmts {
			switch s := stmt.(type) {
			case *ast.IfStmt:
				// Close previous block
				condID := fmt.Sprintf("B_if_cond_%d", blockCounter)
				thenID := fmt.Sprintf("B_then_%d", blockCounter)
				elseID := fmt.Sprintf("B_else_%d", blockCounter)
				doneID := fmt.Sprintf("B_if_done_%d", blockCounter)
				blockCounter++

				blocks = append(blocks, Block{ID: condID, Type: "if_cond", Label: "Check conditional statement"})
				edges = append(edges, CFGEdge{From: currBlockID, To: condID, Label: "next"})

				// Then branch
				blocks = append(blocks, Block{ID: thenID, Type: "then", Label: "Then execution block"})
				edges = append(edges, CFGEdge{From: condID, To: thenID, Label: "true"})

				// Build Then CFG recursively
				prevBlock := currBlockID
				currBlockID = thenID
				if s.Body != nil {
					walkStatements(s.Body.List)
				}
				edges = append(edges, CFGEdge{From: currBlockID, To: doneID, Label: "next"})

				// Else branch
				currBlockID = condID
				if s.Else != nil {
					blocks = append(blocks, Block{ID: elseID, Type: "else", Label: "Else execution block"})
					edges = append(edges, CFGEdge{From: condID, To: elseID, Label: "false"})

					currBlockID = elseID
					switch elseStmt := s.Else.(type) {
					case *ast.BlockStmt:
						walkStatements(elseStmt.List)
					case *ast.IfStmt:
						walkStatements([]ast.Stmt{elseStmt})
					}
					edges = append(edges, CFGEdge{From: currBlockID, To: doneID, Label: "next"})
				} else {
					edges = append(edges, CFGEdge{From: condID, To: doneID, Label: "false"})
				}

				// Convergence
				blocks = append(blocks, Block{ID: doneID, Type: "done", Label: "Converge conditionals"})
				currBlockID = doneID
				_ = prevBlock

			case *ast.ForStmt, *ast.RangeStmt:
				headerID := fmt.Sprintf("B_loop_hdr_%d", blockCounter)
				bodyID := fmt.Sprintf("B_loop_body_%d", blockCounter)
				doneID := fmt.Sprintf("B_loop_done_%d", blockCounter)
				blockCounter++

				blocks = append(blocks, Block{ID: headerID, Type: "loop_header", Label: "Loop check boundary"})
				edges = append(edges, CFGEdge{From: currBlockID, To: headerID, Label: "next"})

				blocks = append(blocks, Block{ID: bodyID, Type: "loop_body", Label: "Loop payload block"})
				edges = append(edges, CFGEdge{From: headerID, To: bodyID, Label: "iterate"})

				currBlockID = bodyID
				if forS, ok := s.(*ast.ForStmt); ok && forS.Body != nil {
					walkStatements(forS.Body.List)
				} else if rangeS, ok := s.(*ast.RangeStmt); ok && rangeS.Body != nil {
					walkStatements(rangeS.Body.List)
				}
				edges = append(edges, CFGEdge{From: currBlockID, To: headerID, Label: "loop-back"})

				blocks = append(blocks, Block{ID: doneID, Type: "done", Label: "Exit loop boundary"})
				edges = append(edges, CFGEdge{From: headerID, To: doneID, Label: "done"})
				currBlockID = doneID

			default:
				stmtCount++
				// Represent sequential blocks easily
				if stmtCount%5 == 0 {
					newID := fmt.Sprintf("B_stmt_%d", blockCounter)
					blockCounter++
					blocks = append(blocks, Block{ID: newID, Type: "stmt", Label: "Statements cluster"})
					edges = append(edges, CFGEdge{From: currBlockID, To: newID, Label: "next"})
					currBlockID = newID
				}
			}
		}
	}

	walkStatements(decl.Body.List)

	blocks = append(blocks, Block{ID: "B_exit", Type: "done", Label: "Exit: Function complete"})
	edges = append(edges, CFGEdge{From: currBlockID, To: "B_exit", Label: "return"})

	cfg.Blocks = blocks
	cfg.Edges = edges
	return cfg
}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func formatSelectorCall(expr *ast.SelectorExpr) string {
	switch x := expr.X.(type) {
	case *ast.Ident:
		return x.Name + "." + expr.Sel.Name
	case *ast.SelectorExpr:
		return formatSelectorCall(x) + "." + expr.Sel.Name
	default:
		return "expr." + expr.Sel.Name
	}
}
