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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/superlinkx/project-lucent/internal/analyzer"
	"github.com/superlinkx/project-lucent/internal/exporter"
)

func main() {
	dirFlag := flag.String("dir", ".", "Root directory of the Go module to scan")
	entriesFlag := flag.String("entry", "", "Comma-separated list of top-level function names (e.g., 'main,api.Handler')")
	formatFlag := flag.String("format", "analysis", "Output format: 'json' (OpenGraph Schema), 'c4' (LikeC4 DSL), 'html' (Interactive Dashboard), 'analysis' (CLI report), 'all'")
	outFlag := flag.String("out", "", "Output file path (prints to stdout if omitted)")
	routesFlag := flag.String("routes", "", "Optional path to a JSON file containing route mapping (handler_id -> route info)")

	flag.Parse()

	if *entriesFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: At least one entry-point function (-entry) is required.")
		flag.Usage()
		os.Exit(1)
	}

	entryPoints := strings.Split(*entriesFlag, ",")
	for i := range entryPoints {
		entryPoints[i] = strings.TrimSpace(entryPoints[i])
	}

	// Parse routes file if specified
	var routeMap map[string]exporter.RouteInfo
	if *routesFlag != "" {
		data, err := os.ReadFile(*routesFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read routes file %s: %v\n", *routesFlag, err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &routeMap); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse routes file %s: %v\n", *routesFlag, err)
			os.Exit(1)
		}
		fmt.Printf("Loaded %d route mappings from %s\n", len(routeMap), *routesFlag)
	}

	fmt.Printf("Initializing Project Lucent analyzer on directory: %s...\n", *dirFlag)
	ma, err := analyzer.NewModuleAnalyzer(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Scanning codebase (Module: %s)...\n", ma.ModulePath)
	if err := ma.Scan(); err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building call graph for entry points: %v...\n", entryPoints)
	graph := ma.Analyze(entryPoints)
	if len(graph) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: Call graph is empty. Ensure entry point names match declared functions in the module.")
	}

	switch *formatFlag {
	case "json":
		data, err := exporter.ExportOpenGraph(graph, routeMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			os.Exit(1)
		}
		writeOutput(data, *outFlag)

	case "c4":
		c4DSL := exporter.ExportLikeC4(graph, routeMap)
		writeOutput([]byte(c4DSL), *outFlag)

	case "html":
		htmlContent, err := exporter.ExportHtml(graph, routeMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			os.Exit(1)
		}
		writeOutput([]byte(htmlContent), *outFlag)

	case "analysis":
		report := RunAnalysisReport(graph, entryPoints, routeMap)
		writeOutput([]byte(report), *outFlag)

	case "all":
		// Export all formats into default files and print report
		jsonData, _ := exporter.ExportOpenGraph(graph, routeMap)
		writeOutput(jsonData, "lucent_opengraph.json")
		fmt.Println("Exported OpenGraph JSON to: lucent_opengraph.json")

		c4DSL := exporter.ExportLikeC4(graph, routeMap)
		writeOutput([]byte(c4DSL), "lucent_diagram.c4")
		fmt.Println("Exported LikeC4 DSL to: lucent_diagram.c4")

		htmlContent, _ := exporter.ExportHtml(graph, routeMap)
		writeOutput([]byte(htmlContent), "lucent_report.html")
		fmt.Println("Exported Interactive Dashboard to: lucent_report.html")

		report := RunAnalysisReport(graph, entryPoints, routeMap)
		fmt.Println("\n--- Analysis Report ---")
		fmt.Print(report)

	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *formatFlag)
		os.Exit(1)
	}
}

func writeOutput(data []byte, filepath string) {
	if filepath == "" {
		fmt.Println(string(data))
		return
	}
	err := os.WriteFile(filepath, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output to file %s: %v\n", filepath, err)
		os.Exit(1)
	}
	fmt.Printf("Output successfully written to: %s\n", filepath)
}

type EntryRanking struct {
	EntryID          string
	FunctionCount    int
	RolledComplexity int
	ReachableSet     map[string]bool
}

func RunAnalysisReport(graph map[string]*analyzer.FunctionNode, entryPoints []string, routeMap map[string]exporter.RouteInfo) string {
	var sb strings.Builder

	sb.WriteString("====================================================\n")
	sb.WriteString("            PROJECT LUCENT ANALYSIS REPORT          \n")
	sb.WriteString("====================================================\n\n")

	// Calculate Rolled Up Cyclomatic Complexity for each entry point and identify its reachable set
	rankings := []EntryRanking{}
	for _, ep := range entryPoints {
		// Locate the full node ID
		fullID := ""
		for id, node := range graph {
			if node.IsEntry && (id == ep || node.Name == ep || strings.HasSuffix(id, "."+ep)) {
				fullID = id
				break
			}
		}

		if fullID == "" {
			continue
		}

		// Compute transitive reachable set using BFS
		visited := make(map[string]bool)
		queue := []string{fullID}
		rolledComplexity := 0

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if visited[curr] {
				continue
			}
			visited[curr] = true

			node := graph[curr]
			if node != nil {
				rolledComplexity += node.Complexity
				for _, child := range node.Calls {
					if !visited[child] {
						queue = append(queue, child)
					}
				}
				for _, boundary := range node.BoundaryCalls {
					visited[boundary] = true
				}
			}
		}

		rankings = append(rankings, EntryRanking{
			EntryID:          fullID,
			FunctionCount:    len(visited),
			RolledComplexity: rolledComplexity,
			ReachableSet:     visited,
		})
	}

	// Group rankings by group name
	groups := make(map[string][]EntryRanking)
	for _, r := range rankings {
		groupName := "Ungrouped"
		if routeMap != nil {
			if info, exists := routeMap[r.EntryID]; exists && info.Group != "" {
				groupName = info.Group
			}
		}
		groups[groupName] = append(groups[groupName], r)
	}

	// Sort group names so that the report is deterministic
	var groupNames []string
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, gName := range groupNames {
		groupRankings := groups[gName]

		// Sort entry points in this group ascending by rolled complexity (simplest to most complex)
		sort.Slice(groupRankings, func(i, j int) bool {
			if groupRankings[i].RolledComplexity == groupRankings[j].RolledComplexity {
				return groupRankings[i].EntryID < groupRankings[j].EntryID
			}
			return groupRankings[i].RolledComplexity < groupRankings[j].RolledComplexity
		})

		sb.WriteString(fmt.Sprintf("ROUTE GROUP: %s\n", strings.ToUpper(gName)))
		sb.WriteString(strings.Repeat("-", len(gName)+13) + "\n")
		sb.WriteString("Rankings (sorted from simplest to most complex):\n")
		for idx, r := range groupRankings {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", idx+1, r.EntryID))
			sb.WriteString(fmt.Sprintf("     - Transitive Function Nodes: %d\n", r.FunctionCount))
			sb.WriteString(fmt.Sprintf("     - Combined Rolled Complexity: %d\n", r.RolledComplexity))
		}
		sb.WriteString("\n")

		// Identify Shared Dependencies within this specific group
		sharedDeps := make(map[string][]string) // funcID -> []entryIDs
		for _, r := range groupRankings {
			for funcID := range r.ReachableSet {
				// Exclude the entry point itself to focus on common shared dependencies
				if funcID != r.EntryID {
					sharedDeps[funcID] = append(sharedDeps[funcID], r.EntryID)
				}
			}
		}

		// Group the shared functions by their exact caller sets
		callerSets := make(map[string][]string) // joined string of callers -> []funcIDs
		for funcID, callers := range sharedDeps {
			if len(callers) > 1 {
				sort.Strings(callers)
				key := strings.Join(callers, ", ")
				callerSets[key] = append(callerSets[key], funcID)
			}
		}

		type CallerSetCluster struct {
			CallersKey   string
			CallersCount int
			Funcs        []string
		}

		clusters := []CallerSetCluster{}
		for key, funcs := range callerSets {
			sort.Strings(funcs)
			callers := strings.Split(key, ", ")
			clusters = append(clusters, CallerSetCluster{
				CallersKey:   key,
				CallersCount: len(callers),
				Funcs:        funcs,
			})
		}

		// Sort clusters:
		// 1. By CallersCount descending (widely-shared first)
		// 2. By number of shared functions descending
		// 3. Alphabetically by CallersKey
		sort.Slice(clusters, func(i, j int) bool {
			if clusters[i].CallersCount != clusters[j].CallersCount {
				return clusters[i].CallersCount > clusters[j].CallersCount
			}
			if len(clusters[i].Funcs) != len(clusters[j].Funcs) {
				return len(clusters[i].Funcs) > len(clusters[j].Funcs)
			}
			return clusters[i].CallersKey < clusters[j].CallersKey
		})

		sb.WriteString("Shared Dependencies / Common Functions:\n")
		if len(clusters) == 0 {
			if len(groupRankings) <= 1 {
				sb.WriteString("  (Only 1 entry point in this group, no shared dependencies possible)\n")
			} else {
				sb.WriteString("  No shared dependencies found across these entry points.\n")
			}
		} else {
			for _, cluster := range clusters {
				sb.WriteString(fmt.Sprintf("  * Shared by %d endpoints: [%s]\n", cluster.CallersCount, cluster.CallersKey))
				for _, f := range cluster.Funcs {
					sb.WriteString(fmt.Sprintf("    - %s\n", f))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
