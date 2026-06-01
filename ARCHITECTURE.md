ARCHITECTURE.md - Project Lucent
Project Lucent is a zero-dependency, high-performance static analysis tool written in Go. Its goal is to trace call graphs starting from a list of entry-point functions within a Go module, calculate cyclomatic complexity for each function node, and export these graphs into collaborative and analytical formats (OpenGraph JSON and LikeC4 diagrams).


1. Technical Design Overview
```text
                                  +-----------------------+
                                  |     Go Source Code    |
                                  +-----------+-----------+
                                              |
                                              v (go/parser & go/ast)
                                  +-----------------------+
                                  |     Module Parser     |
                                  +-----------+-----------+
                                              |
                                              v
                                  +-----------------------+
                                  |      Call Graph       |
                                  |      Constructor      |
                                  +-----------+-----------+
                                              |
               +------------------------------+------------------------------+
               v                              v                              v
   +-------------------------+    +-------------------------+    +-------------------------+
   |  Complexity Analyzer    |    |   OpenGraph Exporter    |    |    LikeC4 Exporter      |
   |  (AST-based complexity) |    |  (OpenGraph JSON format)|   |  (.c4 DSL generation)   |
   +-------------------------+    +-------------------------+    +-------------------------+
```


2. Component Specifications
2.1. Module Parser & Call Graph Constructor
To adhere to our zero-dependency goal, we use the standard library go/token, go/parser, and go/ast to construct the call graph.

1. Module Discovery: Read the local go.mod file to determine the root module name (e.g., github.com/user/lucent). This path acts as the boundary: any imported package starting with this prefix is internal; all others are external and represent traversal endpoints.
2. Function Registry: Parse all .go files recursively in the targeted module directories, building a map of fully qualified function names to their AST representations (*ast.FuncDecl).
3. Call Resolution:
  - For each function declaration, use ast.Inspect to find all ast.CallExpr (call sites).
  - Local Calls: If the call is an identifier (e.g. helper()), resolve it to current_package.helper.
  - Imported Calls: If the call is a selector expression (e.g. pkg.Func()), map the selector to the file's imports to find the imported package.
  - External Calls: If the package prefix does not match the local module path, classify it as an external dependency and stop traversal.
4. Graph Traversal: Beginning with a list of target entry-points, traverse the call tree using a Depth-First Search (DFS) or Breadth-First Search (BFS) to construct a directed graph.

Design Note: Pure static AST parsing resolves direct package functions and struct method calls. Interfaces and dynamic dispatch require abstract interpretation or SSA, which can be explored in later milestones.


2.2. Cyclomatic Complexity Analyzer
We implement an AST-based complexity analyzer using the standard formula Complexity = decision_points + 1.

For each function's AST subtree, we walk the nodes and increment the score by 1 for each of the following:

- *ast.IfStmt (Conditional branches)
- *ast.ForStmt (Loops)
- *ast.RangeStmt (Range-based loops)
- *ast.CaseClause (Switch-case branches, excluding default)
- *ast.CommClause (Select-case branches, excluding default)
- *ast.BinaryExpr where the operator is logical && (token.LAND) or || (token.LOR)

This complexity score is stored as a node property.


2.3. Exporters
A. OpenGraph Exporter
Conforms to the OpenGraph JSON schema.

```json
{
  "graph": {
    "nodes": [
      {
        "id": "github.com/user/lucent/pkg/analyzer.Analyze",
        "kinds": ["Function"],
        "properties": {
          "name": "pkg/analyzer.Analyze",
          "displayname": "Analyze",
          "cyclomatic_complexity": 5,
          "is_entry": false
        }
      }
    ],
    "edges": [
      {
        "start_node_id": "github.com/user/lucent/main.main",
        "end_node_id": "github.com/user/lucent/pkg/analyzer.Analyze",
        "kind": "Calls",
        "properties": {}
      }
    ]
  },
  "metadata": {
    "source_kind": "ProjectLucent"
  }
}
```

B. LikeC4 Diagram Exporter
Generates standard .c4 DSL structure representing packages as parent containers and functions as nested components.

```c4
specification {
  element package {
    style {
      shape rectangle
    }
  }
  element function {
    style {
      shape rectangle
    }
  }
}

model {
  pkg_main = package 'main' {
    func_main = function 'main()' {
      description 'Cyclomatic Complexity: 1'
    }
  }

  pkg_analyzer = package 'analyzer' {
    func_analyze = function 'Analyze()' {
      description 'Cyclomatic Complexity: 5'
    }
  }

  func_main -> func_analyze 'calls'
}

views {
  view index {
    title 'Project Lucent Call Graph'
    include *
  }
}
```

2.4. Control Flow Graph (CFG) Integration
We can construct a Control Flow Graph (CFG) for each function to model internal paths. This can be achieved using the official extended-standard library golang.org/x/tools/go/cfg package, or by modeling a custom AST block walker to remain strictly standard library-only. The CFG is represented as a series of basic blocks and execution transitions, which is then stored as a structured JSON property ('cfg') on the Function node. This CFG can be exported in DOT format or modeled as detailed sub-views and nested components in LikeC4, allowing developers to visually trace the internal paths of highly complex functions.
3. Analysis Engines
1. Rolled Up Complexity Ranking:

  - For any entry point function, transitively traverse all reachable internal nodes.
  - Sum the cyclomatic complexities of all unique nodes in this set to produce the "Rolled Up Complexity" score.
  - Sort entry points to rank complexity from simplest to most demanding.

2. Shared Dependency Identification:

  - Traverses call graphs for all entry points and flags nodes that are reachable from more than one entry point.
  - Returns a structured mapping: SharedFunction -> []EntryPoints. This identifies parallelization bottlenecks and high-risk modification targets.
