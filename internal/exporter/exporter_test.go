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

package exporter_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/lucent/internal/analyzer"
	"github.com/user/lucent/internal/exporter"
)

func TestExportOpenGraph(t *testing.T) {
	hashID := func(id string) string {
		h := sha256.Sum256([]byte(id))
		return hex.EncodeToString(h[:])
	}

	funcs := map[string]*analyzer.FunctionNode{
		"main.main": {
			ID:            "main.main",
			Package:       "main",
			Name:          "main",
			Complexity:    1,
			Calls:         []string{"internal/analyzer.NewModuleAnalyzer"},
			ExternalCalls: []string{"fmt.Println"},
			IsEntry:       true,
			CFG: analyzer.CFG{
				Blocks: []analyzer.Block{
					{ID: "B_entry", Type: "entry", Label: "Entry: main"},
					{ID: "B_exit", Type: "done", Label: "Exit"},
				},
				Edges: []analyzer.CFGEdge{
					{From: "B_entry", To: "B_exit", Label: "return"},
				},
			},
		},
		"internal/analyzer.NewModuleAnalyzer": {
			ID:         "internal/analyzer.NewModuleAnalyzer",
			Package:    "internal/analyzer",
			Name:       "NewModuleAnalyzer",
			Complexity: 3,
			Calls:      []string{},
			CFG: analyzer.CFG{
				Blocks: []analyzer.Block{
					{ID: "B_entry", Type: "entry", Label: "Entry: NewModuleAnalyzer"},
				},
			},
		},
	}

	data, err := exporter.ExportOpenGraph(funcs, nil)
	require.NoError(t, err)

	// Since we verify the structured JSON, unmarshal it back
	type LocalOpenGraphNode struct {
		ID         string                 `json:"id"`
		Kinds      []string               `json:"kinds"`
		Properties map[string]interface{} `json:"properties"`
	}
	type LocalEdgeEndpoint struct {
		Value   string `json:"value"`
		Kind    string `json:"kind"`
		MatchBy string `json:"match_by"`
	}
	type LocalOpenGraphEdge struct {
		Start      LocalEdgeEndpoint      `json:"start"`
		End        LocalEdgeEndpoint      `json:"end"`
		Kind       string                 `json:"kind"`
		Properties map[string]interface{} `json:"properties,omitempty"`
	}
	type LocalOpenGraphGraph struct {
		Nodes []LocalOpenGraphNode `json:"nodes"`
		Edges []LocalOpenGraphEdge `json:"edges"`
	}
	type LocalPayload struct {
		Graph    LocalOpenGraphGraph    `json:"graph"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	var payload LocalPayload
	err = json.Unmarshal(data, &payload)
	require.NoError(t, err)

	assert.Equal(t, "ProjectLucent", payload.Metadata["source_kind"])
	// 2 function nodes + 3 basic block nodes = 5 nodes
	assert.Len(t, payload.Graph.Nodes, 5)
	// 1 call edge + 1 flow edge + 2 hasCFG edges = 4 edges
	assert.Len(t, payload.Graph.Edges, 4)

	// Validate nodes
	var mainNode *LocalOpenGraphNode
	var blockNode *LocalOpenGraphNode
	for i := range payload.Graph.Nodes {
		if payload.Graph.Nodes[i].ID == hashID("main.main") {
			mainNode = &payload.Graph.Nodes[i]
		}
		if payload.Graph.Nodes[i].ID == hashID("main.main:B_entry") {
			blockNode = &payload.Graph.Nodes[i]
		}
	}
	require.NotNil(t, mainNode)
	assert.Equal(t, "main", mainNode.Properties["displayname"])
	assert.Equal(t, float64(1), mainNode.Properties["cyclomatic_complexity"])
	assert.Equal(t, true, mainNode.Properties["is_entry"])
	assert.Nil(t, mainNode.Properties["cfg"], "CFG should not be embedded as a property")

	require.NotNil(t, blockNode)
	assert.Equal(t, []string{"BasicBlock"}, blockNode.Kinds)
	assert.Equal(t, "entry", blockNode.Properties["type"])
	assert.Equal(t, "Entry: main", blockNode.Properties["label"])
	assert.Equal(t, "main.main", blockNode.Properties["function_id"])

	// Validate edges
	var callEdge *LocalOpenGraphEdge
	var flowEdge *LocalOpenGraphEdge
	var cfgEntryEdge *LocalOpenGraphEdge
	for i := range payload.Graph.Edges {
		edge := &payload.Graph.Edges[i]
		if edge.Kind == "Calls" {
			callEdge = edge
		} else if edge.Kind == "FlowsTo" {
			flowEdge = edge
		} else if edge.Kind == "HasCFG" && edge.Start.Value == hashID("main.main") {
			cfgEntryEdge = edge
		}
	}

	require.NotNil(t, callEdge)
	assert.Equal(t, hashID("main.main"), callEdge.Start.Value)
	assert.Equal(t, "Function", callEdge.Start.Kind)
	assert.Equal(t, hashID("internal/analyzer.NewModuleAnalyzer"), callEdge.End.Value)
	assert.Equal(t, "Function", callEdge.End.Kind)

	require.NotNil(t, flowEdge)
	assert.Equal(t, hashID("main.main:B_entry"), flowEdge.Start.Value)
	assert.Equal(t, "BasicBlock", flowEdge.Start.Kind)
	assert.Equal(t, hashID("main.main:B_exit"), flowEdge.End.Value)
	assert.Equal(t, "BasicBlock", flowEdge.End.Kind)
	assert.Equal(t, "return", flowEdge.Properties["label"])

	require.NotNil(t, cfgEntryEdge)
	assert.Equal(t, hashID("main.main:B_entry"), cfgEntryEdge.End.Value)
	assert.Equal(t, "BasicBlock", cfgEntryEdge.End.Kind)
}

func TestExportLikeC4(t *testing.T) {
	funcs := map[string]*analyzer.FunctionNode{
		"main.main": {
			ID:         "main.main",
			Package:    "main",
			Name:       "main",
			Complexity: 1,
			Calls:      []string{"internal/analyzer.NewModuleAnalyzer"},
		},
		"internal/analyzer.NewModuleAnalyzer": {
			ID:         "internal/analyzer.NewModuleAnalyzer",
			Package:    "internal/analyzer",
			Name:       "NewModuleAnalyzer",
			Complexity: 3,
			Calls:      []string{},
		},
	}

	dsl := exporter.ExportLikeC4(funcs, nil)

	// Verify LikeC4 syntax structure and escaped package/function identifiers
	assert.Contains(t, dsl, "specification {")
	assert.Contains(t, dsl, "model {")
	assert.Contains(t, dsl, "views {")
	assert.Contains(t, dsl, "main = package 'main'")
	assert.Contains(t, dsl, "internal_analyzer = package 'internal/analyzer'")
	assert.Contains(t, dsl, "main_main = function 'main()'")
	assert.Contains(t, dsl, "internal_analyzer_NewModuleAnalyzer = function 'NewModuleAnalyzer()'")
	assert.Contains(t, dsl, "main_main -> internal_analyzer_NewModuleAnalyzer 'calls'")
}
