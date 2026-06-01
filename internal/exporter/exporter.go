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

package exporter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/superlinkx/project-lucent/internal/analyzer"
)

// OpenGraphNode represents a node in the OpenGraph schema.
type OpenGraphNode struct {
	ID         string                 `json:"id"`
	Kinds      []string               `json:"kinds"`
	Properties map[string]interface{} `json:"properties"`
}

// EdgeEndpoint represents a match endpoint in the OpenGraph ingest schemas.
type EdgeEndpoint struct {
	Value   string `json:"value"`
	Kind    string `json:"kind"`
	MatchBy string `json:"match_by"`
}

// OpenGraphEdge represents a relationship edge in the OpenGraph schema.
type OpenGraphEdge struct {
	Start      EdgeEndpoint           `json:"start"`
	End        EdgeEndpoint           `json:"end"`
	Kind       string                 `json:"kind"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// OpenGraphGraph holds the nodes and edges arrays.
type OpenGraphGraph struct {
	Nodes []OpenGraphNode `json:"nodes"`
	Edges []OpenGraphEdge `json:"edges"`
}

// RouteInfo represents routing metadata for an API endpoint.
type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Group  string `json:"group"`
}

// OpenGraphPayload represents the top level OpenGraph JSON payload.
type OpenGraphPayload struct {
	Graph    OpenGraphGraph         `json:"graph"`
	Metadata map[string]interface{} `json:"metadata"`
}

// hashID returns a stable lowercase hex representation of the SHA-256 hash of the given ID.
// This ensures IDs are case-insensitive and safe from case-collisions/truncation in databases.
func hashID(id string) string {
	h := sha256.Sum256([]byte(id))
	return hex.EncodeToString(h[:])
}

// ExportOpenGraph serializes the analyzer function map to OpenGraph JSON format.
func ExportOpenGraph(funcs map[string]*analyzer.FunctionNode, routeMap map[string]RouteInfo) ([]byte, error) {
	nodes := []OpenGraphNode{}
	edges := []OpenGraphEdge{}

	for id, f := range funcs {
		// 1. Add Function Node (without "cfg" property)
		properties := map[string]interface{}{
			"name":                  id,
			"displayname":           f.Name,
			"package":               f.Package,
			"cyclomatic_complexity": f.Complexity,
			"is_entry":              f.IsEntry,
		}

		if routeMap != nil {
			if rInfo, exists := routeMap[id]; exists {
				properties["route_path"] = rInfo.Path
				properties["route_method"] = rInfo.Method
				properties["route_group"] = rInfo.Group
			}
		}

		nodes = append(nodes, OpenGraphNode{
			ID:         hashID(id),
			Kinds:      []string{"Function"},
			Properties: properties,
		})

		// 2. Add calling edges to other functions
		for _, childID := range f.Calls {
			edges = append(edges, OpenGraphEdge{
				Start: EdgeEndpoint{
					Value:   hashID(id),
					Kind:    "Function",
					MatchBy: "id",
				},
				End: EdgeEndpoint{
					Value:   hashID(childID),
					Kind:    "Function",
					MatchBy: "id",
				},
				Kind: "Calls",
			})
		}

		// 3. Add CFG Blocks as Nodes
		for _, block := range f.CFG.Blocks {
			blockNodeID := fmt.Sprintf("%s:%s", id, block.ID)
			nodes = append(nodes, OpenGraphNode{
				ID:    hashID(blockNodeID),
				Kinds: []string{"BasicBlock"},
				Properties: map[string]interface{}{
					"name":        blockNodeID,
					"displayname": block.ID,
					"type":        block.Type,
					"label":       block.Label,
					"function_id": id,
				},
			})
		}

		// 4. Add CFG Edges as Edges
		for _, edge := range f.CFG.Edges {
			edges = append(edges, OpenGraphEdge{
				Start: EdgeEndpoint{
					Value:   hashID(fmt.Sprintf("%s:%s", id, edge.From)),
					Kind:    "BasicBlock",
					MatchBy: "id",
				},
				End: EdgeEndpoint{
					Value:   hashID(fmt.Sprintf("%s:%s", id, edge.To)),
					Kind:    "BasicBlock",
					MatchBy: "id",
				},
				Kind: "FlowsTo",
				Properties: map[string]interface{}{
					"label": edge.Label,
				},
			})
		}

		// 5. Connect Function node to its Entry basic block node
		edges = append(edges, OpenGraphEdge{
			Start: EdgeEndpoint{
				Value:   hashID(id),
				Kind:    "Function",
				MatchBy: "id",
			},
			End: EdgeEndpoint{
				Value:   hashID(fmt.Sprintf("%s:B_entry", id)),
				Kind:    "BasicBlock",
				MatchBy: "id",
			},
			Kind: "HasCFG",
		})
	}

	payload := OpenGraphPayload{
		Graph: OpenGraphGraph{
			Nodes: nodes,
			Edges: edges,
		},
		Metadata: map[string]interface{}{
			"source_kind": "ProjectLucent",
		},
	}

	return json.MarshalIndent(payload, "", "  ")
}

// ExportLikeC4 translates the analysis graph into a valid LikeC4 architecture DSL block.
func ExportLikeC4(funcs map[string]*analyzer.FunctionNode, routeMap map[string]RouteInfo) string {
	var sb strings.Builder

	sb.WriteString("specification {\n")
	sb.WriteString("  element package {\n")
	sb.WriteString("    style {\n")
	sb.WriteString("      shape rectangle\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("  element function {\n")
	sb.WriteString("    style {\n")
	sb.WriteString("      shape rectangle\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")

	sb.WriteString("model {\n")

	// Group functions by package
	pkgGroups := make(map[string][]*analyzer.FunctionNode)
	for _, f := range funcs {
		pkgGroups[f.Package] = append(pkgGroups[f.Package], f)
	}

	// Generate nodes grouped by package container
	for pkg, pkgFuncs := range pkgGroups {
		escapedPkgID := escapeID(pkg)
		sb.WriteString(fmt.Sprintf("  %s = package '%s' {\n", escapedPkgID, pkg))

		for _, f := range pkgFuncs {
			escapedFuncID := escapeID(f.ID)
			sb.WriteString(fmt.Sprintf("    %s = function '%s()' {\n", escapedFuncID, f.Name))
			
			desc := fmt.Sprintf("Cyclomatic Complexity: %d", f.Complexity)
			if routeMap != nil {
				if rInfo, exists := routeMap[f.ID]; exists {
					desc += fmt.Sprintf(" | Route: %s %s", rInfo.Method, rInfo.Path)
				}
			}
			sb.WriteString(fmt.Sprintf("      description '%s'\n", desc))
			sb.WriteString("    }\n")
		}

		sb.WriteString("  }\n")
	}

	sb.WriteString("\n  // Relationships\n")
	// Generate relationship edges
	for id, f := range funcs {
		escapedFrom := escapeID(id)
		for _, callID := range f.Calls {
			if _, exists := funcs[callID]; exists {
				escapedTo := escapeID(callID)
				sb.WriteString(fmt.Sprintf("  %s -> %s 'calls'\n", escapedFrom, escapedTo))
			}
		}
	}

	sb.WriteString("}\n\n")

	sb.WriteString("views {\n")
	sb.WriteString("  view index {\n")
	sb.WriteString("    title 'Project Lucent - Call Graph'\n")
	sb.WriteString("    include *\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// escapeID removes or replaces special characters in fully qualified paths
// so they form valid LikeC4 identifiers (alphanumeric, underscores, hyphens).
func escapeID(raw string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	escaped := re.ReplaceAllString(raw, "_")
	// Clean up consecutive underscores
	for strings.Contains(escaped, "__") {
		escaped = strings.ReplaceAll(escaped, "__", "_")
	}
	escaped = strings.Trim(escaped, "_")
	if len(escaped) > 0 && escaped[0] >= '0' && escaped[0] <= '9' {
		escaped = "node_" + escaped
	}
	return escaped
}

// ExportHtml generates a self-contained interactive HTML dashboard from the call graph.
func ExportHtml(graph map[string]*analyzer.FunctionNode, routeMap map[string]RouteInfo) (string, error) {
	graphBytes, err := json.Marshal(graph)
	if err != nil {
		return "", err
	}

	routeMapBytes, err := json.Marshal(routeMap)
	if err != nil {
		return "", err
	}

	htmlContent := strings.Replace(htmlTemplate, "{{.GraphData}}", string(graphBytes), 1)
	htmlContent = strings.Replace(htmlContent, "{{.RouteMapData}}", string(routeMapBytes), 1)

	return htmlContent, nil
}

