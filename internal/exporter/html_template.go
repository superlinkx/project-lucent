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

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Project Lucent - Interactive Call Graph Dashboard</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Outfit:wght@500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-main: #0b0f19;
      --bg-sidebar: #111625;
      --bg-card: rgba(22, 28, 45, 0.6);
      --border-card: rgba(255, 255, 255, 0.08);
      --border-focus: #4f46e5;
      --text-main: #f3f4f6;
      --text-muted: #9ca3af;
      --text-dim: #6b7280;
      
      --accent-primary: #6366f1;
      --accent-primary-glow: rgba(99, 102, 241, 0.15);
      --accent-teal: #14b8a6;
      
      --method-get: #10b981;
      --method-post: #3b82f6;
      --method-put: #f59e0b;
      --method-delete: #ef4444;
      --method-any: #8b5cf6;

      --font-sans: 'Inter', sans-serif;
      --font-title: 'Outfit', sans-serif;
      --font-mono: 'JetBrains Mono', monospace;
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      background-color: var(--bg-main);
      background-image: 
        radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.1) 0px, transparent 50%),
        radial-gradient(at 100% 100%, rgba(20, 184, 166, 0.05) 0px, transparent 50%);
      color: var(--text-main);
      font-family: var(--font-sans);
      min-block-size: 100vh;
      overflow-x: hidden;
      display: grid;
      grid-template-columns: 280px 1fr;
    }

    /* Scrollbar Styling */
    ::-webkit-scrollbar {
      width: 8px;
      height: 8px;
    }
    ::-webkit-scrollbar-track {
      background: rgba(0, 0, 0, 0.2);
    }
    ::-webkit-scrollbar-thumb {
      background: rgba(255, 255, 255, 0.1);
      border-radius: 4px;
    }
    ::-webkit-scrollbar-thumb:hover {
      background: rgba(255, 255, 255, 0.2);
    }

    /* Sidebar Styles */
    aside {
      background-color: var(--bg-sidebar);
      border-inline-end: 1px solid var(--border-card);
      padding: 1.5rem;
      display: flex;
      flex-direction: column;
      height: 100vh;
      position: sticky;
      top: 0;
    }

    .logo-container {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-bottom: 2rem;
    }

    .logo-icon {
      width: 32px;
      height: 32px;
      background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-teal) 100%);
      border-radius: 8px;
      display: grid;
      place-content: center;
      font-weight: 800;
      color: white;
      font-family: var(--font-title);
      font-size: 1.1rem;
    }

    .logo-title {
      font-family: var(--font-title);
      font-size: 1.25rem;
      font-weight: 700;
      background: linear-gradient(to right, #ffffff, #9ca3af);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }

    .search-container {
      margin-bottom: 1.5rem;
      position: relative;
    }

    .search-input {
      width: 100%;
      background: rgba(0, 0, 0, 0.3);
      border: 1px solid var(--border-card);
      border-radius: 8px;
      padding: 0.6rem 0.8rem;
      color: var(--text-main);
      font-family: var(--font-sans);
      font-size: 0.875rem;
      outline: none;
      transition: all 0.2s;
    }

    .search-input:focus {
      border-color: var(--border-focus);
      box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.2);
    }

    .group-list-header {
      font-family: var(--font-title);
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-dim);
      margin-bottom: 0.75rem;
    }

    .group-list {
      list-style: none;
      overflow-y: auto;
      flex-grow: 1;
      padding-bottom: 1rem;
    }

    .group-item {
      padding: 0.5rem 0.75rem;
      border-radius: 6px;
      cursor: pointer;
      font-size: 0.875rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
      transition: all 0.2s;
      margin-bottom: 0.25rem;
      color: var(--text-muted);
    }

    .group-item:hover {
      background: rgba(255, 255, 255, 0.03);
      color: var(--text-main);
    }

    .group-item.active {
      background: rgba(99, 102, 241, 0.15);
      border-left: 3px solid var(--accent-primary);
      color: var(--text-main);
      font-weight: 500;
    }

    .group-badge {
      background: rgba(255, 255, 255, 0.08);
      padding: 0.1rem 0.4rem;
      border-radius: 10px;
      font-size: 0.75rem;
      color: var(--text-muted);
    }

    .group-item.active .group-badge {
      background: rgba(99, 102, 241, 0.3);
      color: white;
    }

    /* Main Content Area */
    main {
      padding: 2rem;
      overflow-y: auto;
      max-block-size: 100vh;
    }

    .dashboard-header {
      margin-bottom: 2rem;
    }

    .dashboard-title {
      font-family: var(--font-title);
      font-size: 2rem;
      font-weight: 700;
      margin-bottom: 0.5rem;
    }

    .dashboard-subtitle {
      color: var(--text-muted);
      font-size: 0.95rem;
    }

    /* Glassmorphism Cards */
    .glass-card {
      background: var(--bg-card);
      backdrop-filter: blur(12px);
      border: 1px solid var(--border-card);
      border-radius: 12px;
      padding: 1.5rem;
      box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.25);
      margin-bottom: 1.5rem;
    }

    /* Stats Grid */
    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
      margin-bottom: 2rem;
    }

    .stat-card {
      padding: 1.25rem;
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .stat-val {
      font-family: var(--font-title);
      font-size: 1.75rem;
      font-weight: 700;
      color: white;
    }

    .stat-lbl {
      font-size: 0.8rem;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    /* Dashboard Layout */
    .dashboard-grid {
      display: grid;
      grid-template-columns: 2fr 1.3fr;
      gap: 1.5rem;
      align-items: start;
    }

    .section-title {
      font-family: var(--font-title);
      font-size: 1.25rem;
      font-weight: 600;
      margin-bottom: 1rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    /* Endpoint Card Styles */
    .endpoint-list {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }

    .endpoint-card {
      padding: 1rem;
      cursor: pointer;
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
      position: relative;
    }

    .endpoint-card:hover {
      border-color: rgba(255, 255, 255, 0.15);
      box-shadow: 0 8px 30px 0 var(--accent-primary-glow);
      transform: translateY(-1px);
    }

    .endpoint-card.active {
      border-color: var(--border-focus);
      background: rgba(99, 102, 241, 0.08);
      box-shadow: 0 0 15px 0 var(--accent-primary-glow);
    }

    .endpoint-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .endpoint-route {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      overflow: hidden;
    }

    .method-badge {
      font-size: 0.7rem;
      font-weight: 700;
      padding: 0.15rem 0.4rem;
      border-radius: 4px;
      text-transform: uppercase;
      flex-shrink: 0;
    }

    .method-get { background: rgba(16, 185, 129, 0.15); color: var(--method-get); }
    .method-post { background: rgba(59, 130, 246, 0.15); color: var(--method-post); }
    .method-put { background: rgba(245, 158, 11, 0.15); color: var(--method-put); }
    .method-delete { background: rgba(239, 68, 68, 0.15); color: var(--method-delete); }
    .method-any { background: rgba(139, 92, 246, 0.15); color: var(--method-any); }

    .route-path {
      font-family: var(--font-mono);
      font-size: 0.85rem;
      color: white;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .endpoint-fqn {
      font-family: var(--font-mono);
      font-size: 0.72rem;
      color: var(--text-muted);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    /* Complexity Visualizer Styles */
    .complexity-bar-container {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 0.25rem;
    }

    .complexity-label {
      font-size: 0.75rem;
      color: var(--text-muted);
      width: 140px;
      flex-shrink: 0;
    }

    .complexity-bar-bg {
      background: rgba(255, 255, 255, 0.05);
      height: 8px;
      border-radius: 4px;
      flex-grow: 1;
      overflow: hidden;
    }

    .complexity-bar-fill {
      height: 100%;
      border-radius: 4px;
      transition: width 0.4s ease-out;
      background: linear-gradient(to right, var(--accent-primary), var(--accent-teal));
    }

    .complexity-value {
      font-family: var(--font-mono);
      font-size: 0.75rem;
      font-weight: 600;
      width: 40px;
      text-align: right;
      flex-shrink: 0;
    }

    /* Shared Dependencies & Overlap Styles */
    .overlap-list {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }

    .overlap-card {
      padding: 1rem;
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      border-radius: 8px;
      border: 1px solid var(--border-card);
      background: rgba(0, 0, 0, 0.2);
    }

    .overlap-card.clickable {
      cursor: pointer;
      transition: all 0.2s ease;
    }

    .overlap-card.clickable:hover {
      background: rgba(99, 102, 241, 0.1);
      border-color: rgba(99, 102, 241, 0.3);
    }

    .overlap-card.clickable.active {
      background: rgba(99, 102, 241, 0.2);
      border-color: var(--accent-primary);
      box-shadow: 0 0 10px rgba(99, 102, 241, 0.25);
    }

    .overlap-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 0.8rem;
    }

    .overlap-names {
      font-weight: 500;
      color: white;
      font-size: 0.85rem;
      display: flex;
      flex-direction: column;
      gap: 0.2rem;
    }

    .overlap-badge {
      background: rgba(99, 102, 241, 0.15);
      color: #818cf8;
      font-weight: 700;
      font-size: 0.75rem;
      padding: 0.15rem 0.4rem;
      border-radius: 4px;
    }

    .overlap-details {
      font-size: 0.75rem;
      color: var(--text-muted);
    }

    /* Shared Functions Map */
    .shared-funcs-list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      max-height: 400px;
      overflow-y: auto;
      padding-right: 0.25rem;
    }

    .shared-func-item {
      padding: 0.75rem;
      background: rgba(0, 0, 0, 0.15);
      border: 1px solid var(--border-card);
      border-radius: 6px;
      font-family: var(--font-mono);
      font-size: 0.72rem;
      cursor: pointer;
      transition: all 0.2s;
    }

    .shared-func-item:hover {
      border-color: var(--accent-primary);
      background: rgba(99, 102, 241, 0.05);
    }

    .shared-func-item.active {
      border-color: var(--accent-teal);
      background: rgba(20, 184, 166, 0.1);
      box-shadow: 0 0 10px rgba(20, 184, 166, 0.1);
    }

    .shared-func-header {
      display: flex;
      justify-content: space-between;
      color: var(--text-main);
      margin-bottom: 0.25rem;
      font-weight: 500;
    }

    .shared-func-count {
      color: var(--accent-teal);
      font-weight: 600;
    }

    /* Modal / Slide-over Panel styles */
    .drawer-overlay {
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.5);
      backdrop-filter: blur(4px);
      z-index: 100;
      opacity: 0;
      pointer-events: none;
      transition: opacity 0.3s ease;
    }

    .drawer-overlay.open {
      opacity: 1;
      pointer-events: auto;
    }

    .drawer {
      position: fixed;
      top: 0;
      right: -550px;
      width: 500px;
      height: 100vh;
      background: var(--bg-sidebar);
      border-inline-start: 1px solid var(--border-card);
      z-index: 101;
      box-shadow: -10px 0 30px rgba(0, 0, 0, 0.5);
      transition: right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      display: flex;
      flex-direction: column;
      padding: 1.5rem;
    }

    .drawer.open {
      right: 0;
    }

    .drawer-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1.5rem;
      border-bottom: 1px solid var(--border-card);
      padding-bottom: 1rem;
    }

    .drawer-title {
      font-family: var(--font-title);
      font-size: 1.25rem;
      font-weight: 700;
      color: white;
    }

    .drawer-close {
      background: none;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      font-size: 1.25rem;
      line-height: 1;
    }

    .drawer-close:hover {
      color: white;
    }

    .drawer-body {
      flex-grow: 1;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }

    .drawer-section-title {
      font-family: var(--font-title);
      font-size: 0.9rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-dim);
      margin-bottom: 0.5rem;
    }

    .drawer-list {
      list-style: none;
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }

    .drawer-list-item {
      font-family: var(--font-mono);
      font-size: 0.72rem;
      padding: 0.5rem;
      background: rgba(0, 0, 0, 0.2);
      border-radius: 4px;
      color: var(--text-muted);
      word-break: break-all;
      transition: all 0.2s ease;
    }

    .drawer-list-item.highlight-shared {
      background: rgba(20, 184, 166, 0.15) !important;
      border-color: rgba(20, 184, 166, 0.4) !important;
      color: white !important;
      box-shadow: inset 3px 0 0 var(--accent-teal);
    }
  </style>
</head>
<body>

  <!-- Sidebar -->
  <aside>
    <div class="logo-container">
      <div class="logo-icon">L</div>
      <div class="logo-title">Project Lucent</div>
    </div>
    
    <div class="search-container">
      <input type="search" id="search-box" class="search-input" placeholder="Search endpoints...">
    </div>

    <!-- Noise Filters -->
    <div class="group-list-header">Global Noise Filters</div>
    <div style="margin-bottom: 1.5rem; display: flex; flex-direction: column; gap: 0.5rem;">
      <input type="text" id="filter-input" class="search-input" placeholder="Exclude patterns (comma-separated)..." style="font-size: 0.8rem; padding: 0.5rem;">
      <div style="display: flex; gap: 0.25rem;">
        <button id="btn-preset-helpers" style="flex: 1; font-size: 0.65rem; padding: 0.3rem 0.5rem; background: rgba(99, 102, 241, 0.2); border: 1px solid var(--border-card); border-radius: 4px; color: white; cursor: pointer;">Helpers</button>
        <button id="btn-preset-loggers" style="flex: 1; font-size: 0.65rem; padding: 0.3rem 0.5rem; background: rgba(99, 102, 241, 0.2); border: 1px solid var(--border-card); border-radius: 4px; color: white; cursor: pointer;">Log/Ctx</button>
        <button id="btn-clear-filters" style="flex: 1; font-size: 0.65rem; padding: 0.3rem 0.5rem; background: rgba(239, 68, 68, 0.2); border: 1px solid var(--border-card); border-radius: 4px; color: white; cursor: pointer;">Clear</button>
      </div>
    </div>

    <div class="group-list-header">Route Groups</div>
    <ul class="group-list" id="group-list-ul">
      <!-- Generated dynamically -->
    </ul>
  </aside>

  <!-- Main Content -->
  <main>
    <div class="dashboard-header">
      <h1 class="dashboard-title" id="dashboard-group-title">Route Group: Loading...</h1>
      <p class="dashboard-subtitle">Select a route group from the sidebar to inspect relative complexities and overlap.</p>
    </div>

    <!-- Stats summary card row -->
    <div class="stats-grid">
      <div class="glass-card stat-card">
        <span class="stat-lbl">Endpoints</span>
        <span class="stat-val" id="stat-endpoints-count">0</span>
      </div>
      <div class="glass-card stat-card">
        <span class="stat-lbl">Total Reachable Nodes</span>
        <span class="stat-val" id="stat-nodes-count">0</span>
      </div>
      <div class="glass-card stat-card">
        <span class="stat-lbl">Average Complexity</span>
        <span class="stat-val" id="stat-avg-complexity">0</span>
      </div>
      <div class="glass-card stat-card">
        <span class="stat-lbl">Max Complexity</span>
        <span class="stat-val" id="stat-max-complexity">0</span>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- Left Column: Endpoints list -->
      <div>
        <div class="section-title">
          <span>Endpoints sorted by Rolled Complexity</span>
          <span style="font-size: 0.8rem; font-weight: normal; color: var(--text-muted);" id="filtered-endpoints-text"></span>
        </div>
        <div class="endpoint-list" id="endpoint-list-container">
          <!-- Generated dynamically -->
        </div>
      </div>

      <!-- Right Column: Interference / Overlaps / Shared Functions -->
      <div>
        <!-- Overlap Matrix/Pairs card -->
        <div class="glass-card">
          <div class="section-title">Top Interference / Overlaps</div>
          <p style="font-size: 0.8rem; color: var(--text-muted); margin-bottom: 1rem;">
            Endpoints sharing the highest percentage of their call stack (indicating refactoring interference).
          </p>
          <div class="overlap-list" id="overlap-list-container">
            <!-- Generated dynamically -->
          </div>
        </div>

        <!-- Shared functions card -->
        <div class="glass-card">
          <div class="section-title">Shared Transitive Functions</div>
          <p style="font-size: 0.8rem; color: var(--text-muted); margin-bottom: 1rem;">
            Select a shared function below to highlight all endpoints in this group that call it.
          </p>
          <div class="shared-funcs-list" id="shared-funcs-container">
            <!-- Generated dynamically -->
          </div>
        </div>
      </div>
    </div>
  </main>

  <!-- Slide-over Drawer details panel -->
  <div class="drawer-overlay" id="drawer-overlay" onclick="closeDrawer()"></div>
  <div class="drawer" id="details-drawer">
    <div class="drawer-header">
      <div>
        <div class="drawer-title" id="drawer-endpoint-title">Endpoint Details</div>
        <div class="endpoint-fqn" id="drawer-endpoint-fqn" style="margin-top: 0.25rem;"></div>
      </div>
      <button class="drawer-close" onclick="closeDrawer()">&times;</button>
    </div>
    <div class="drawer-body">
      <div>
        <div class="drawer-section-title">Metrics</div>
        <div style="display: flex; gap: 2rem; margin-top: 0.5rem;">
          <div>
            <div style="font-size: 0.75rem; color: var(--text-muted);">ROLLED COMPLEXITY</div>
            <div style="font-size: 1.5rem; font-weight: 700; color: white;" id="drawer-rolled-cc">0</div>
          </div>
          <div>
            <div style="font-size: 0.75rem; color: var(--text-muted);">TRANSITIVE FUNCTIONS</div>
            <div style="font-size: 1.5rem; font-weight: 700; color: white;" id="drawer-nodes-count">0</div>
          </div>
          <div>
            <div style="font-size: 0.75rem; color: var(--text-muted);">BASE COMPLEXITY</div>
            <div style="font-size: 1.5rem; font-weight: 700; color: white;" id="drawer-base-cc">0</div>
          </div>
        </div>
      </div>

      <div>
        <div class="drawer-section-title">Code Overlaps with Siblings</div>
        <div class="overlap-list" id="drawer-overlaps-list" style="margin-top: 0.5rem; max-height: 200px; overflow-y: auto;">
          <!-- Overlaps generated dynamically -->
        </div>
      </div>

      <div style="display: flex; flex-direction: column; margin-bottom: 1.5rem;">
        <div class="drawer-section-title">Method Call Boundaries (Leafs)</div>
        <div style="max-height: 200px; overflow-y: auto; padding-right: 0.25rem;">
          <ul class="drawer-list" id="drawer-boundaries-list">
            <!-- Dynamic list -->
          </ul>
        </div>
      </div>

      <div style="flex-grow: 1; display: flex; flex-direction: column;">
        <div class="drawer-section-title">Transitive Function Dependencies</div>
        <div style="flex-grow: 1; overflow-y: auto; max-block-size: calc(100vh - 550px); padding-right: 0.25rem;">
          <ul class="drawer-list" id="drawer-dependencies-list">
            <!-- Dynamic list -->
          </ul>
        </div>
      </div>
    </div>
  </div>

  <script>
    // Embedded Data injected by Lucent
    const graph = {{.GraphData}};
    const routeMap = {{.RouteMapData}} || {};

    // Internal State
    let entryNodes = {};
    let groups = {};
    let activeGroup = '';
    let selectedSharedFunc = '';
    let searchQuery = '';
    let activeEntryID = '';

    // Step 1: Precompute transitive reachability (BFS)
    function initializeData() {
      entryNodes = {};
      for (const [id, node] of Object.entries(graph)) {
        if (node.is_entry) {
          const visited = new Set();
          const queue = [id];
          let rolledComplexity = 0;

          while (queue.length > 0) {
            const curr = queue.shift();
            if (visited.has(curr)) continue;
            if (isFilteredOut(curr)) continue;
            visited.add(curr);

            const currNode = graph[curr];
            if (currNode) {
              rolledComplexity += currNode.cyclomatic_complexity;
              for (const child of currNode.calls) {
                if (!visited.has(child) && !isFilteredOut(child)) {
                  queue.push(child);
                }
              }
              if (currNode.boundary_calls) {
                for (const boundary of currNode.boundary_calls) {
                  if (!isFilteredOut(boundary)) {
                    visited.add(boundary);
                  }
                }
              }
            }
          }

          const route = routeMap[id] || { method: 'ANY', path: id, group: 'Ungrouped' };
          
          entryNodes[id] = {
            id,
            name: node.name,
            pkg: node.package,
            complexity: node.cyclomatic_complexity,
            rolledComplexity,
            reachableCount: visited.size,
            reachableSet: visited,
            route
          };
        }
      }

      // Group endpoints by their route.group
      groups = {};
      for (const entry of Object.values(entryNodes)) {
        const gName = entry.route.group || 'Ungrouped';
        if (!groups[gName]) groups[gName] = [];
        groups[gName].push(entry);
      }
    }

    // Step 2: Render sidebar navigation
    function renderSidebar() {
      const ul = document.getElementById('group-list-ul');
      ul.innerHTML = '';

      // Sort group names alphabetically
      const sortedGroupNames = Object.keys(groups).sort();
      
      // Default to first group if none active
      if (!activeGroup && sortedGroupNames.length > 0) {
        activeGroup = sortedGroupNames[0];
      }

      sortedGroupNames.forEach(gName => {
        const li = document.createElement('li');
        li.className = 'group-item' + (activeGroup === gName ? ' active' : '');
        li.innerHTML = '<span>' + gName + '</span><span class="group-badge">' + groups[gName].length + '</span>';
        li.onclick = () => selectGroup(gName);
        ul.appendChild(li);
      });
    }

    // Handle group selection
    function selectGroup(gName) {
      activeGroup = gName;
      selectedSharedFunc = ''; // Reset shared func highlight
      renderSidebar();
      renderGroupDashboard();
    }

    // Step 3: Render dashboard for the active group
    function renderGroupDashboard() {
      if (!activeGroup || !groups[activeGroup]) return;

      document.getElementById('dashboard-group-title').innerText = 'Route Group: ' + activeGroup.toUpperCase();
      
      let groupEntries = [...groups[activeGroup]];

      // Filter by Search Query
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        groupEntries = groupEntries.filter(e => 
          e.route.path.toLowerCase().includes(query) ||
          e.name.toLowerCase().includes(query) ||
          e.id.toLowerCase().includes(query)
        );
      }

      document.getElementById('filtered-endpoints-text').innerText = searchQuery 
        ? 'Found ' + groupEntries.length + ' of ' + groups[activeGroup].length + ' endpoints' 
        : '';

      // Compute statistics for the group (always based on the full unfiltered group)
      const fullGroupEntries = groups[activeGroup];
      const endpointsCount = fullGroupEntries.length;
      
      const allReachableNodes = new Set();
      let totalComplexity = 0;
      let maxComplexity = 0;
      
      fullGroupEntries.forEach(e => {
        totalComplexity += e.rolledComplexity;
        if (e.rolledComplexity > maxComplexity) maxComplexity = e.rolledComplexity;
        e.reachableSet.forEach(nodeID => allReachableNodes.add(nodeID));
      });

      const avgComplexity = Math.round(totalComplexity / endpointsCount);

      document.getElementById('stat-endpoints-count').innerText = endpointsCount;
      document.getElementById('stat-nodes-count').innerText = allReachableNodes.size;
      document.getElementById('stat-avg-complexity').innerText = avgComplexity;
      document.getElementById('stat-max-complexity').innerText = maxComplexity;

      // Sort entries ascending by rolledComplexity
      groupEntries.sort((a, b) => {
        if (a.rolledComplexity === b.rolledComplexity) {
          return a.id.localeCompare(b.id);
        }
        return a.rolledComplexity - b.rolledComplexity;
      });

      // Render Endpoint Cards
      const listContainer = document.getElementById('endpoint-list-container');
      listContainer.innerHTML = '';

      groupEntries.forEach(entry => {
        const card = document.createElement('div');
        card.className = 'glass-card endpoint-card';
        card.dataset.id = entry.id;

        // Visual rolled complexity progress percentage
        const progressPct = maxComplexity > 0 ? (entry.rolledComplexity / maxComplexity) * 100 : 0;
        const methodClass = 'method-' + entry.route.method.toLowerCase();

        card.innerHTML = 
          '<div class="endpoint-header">' +
            '<div class="endpoint-route">' +
              '<span class="method-badge ' + methodClass + '">' + entry.route.method + '</span>' +
              '<span class="route-path" title="' + entry.route.path + '">' + entry.route.path + '</span>' +
            '</div>' +
            '<div style="font-size: 0.75rem; color: var(--text-dim);">Reachable Nodes: ' + entry.reachableCount + '</div>' +
          '</div>' +
          '<div class="endpoint-fqn" title="' + entry.id + '">' + entry.id + '</div>' +
          '<div class="complexity-bar-container">' +
            '<div class="complexity-label">Transitive Rolled Complexity</div>' +
            '<div class="complexity-bar-bg">' +
              '<div class="complexity-bar-fill" style="width: ' + progressPct + '%"></div>' +
            '</div>' +
            '<div class="complexity-value">' + entry.rolledComplexity + '</div>' +
          '</div>';
        card.onclick = () => openEndpointDetails(entry.id);
        listContainer.appendChild(card);
      });

      renderInterferencePairs(groupEntries);
      renderSharedFunctions(fullGroupEntries);
    }

    // Step 4: Render Top Interference Pairs
    function renderInterferencePairs(groupEntries) {
      const container = document.getElementById('overlap-list-container');
      container.innerHTML = '';

      // Compute pairwise overlaps
      const overlaps = [];
      for (let i = 0; i < groupEntries.length; i++) {
        for (let j = i + 1; j < groupEntries.length; j++) {
          const e1 = groupEntries[i];
          const e2 = groupEntries[j];

          // Intersection of reachable function sets
          const intersect = new Set([...e1.reachableSet].filter(x => e2.reachableSet.has(x)));
          if (intersect.size > 0) {
            const minSize = Math.min(e1.reachableSet.size, e2.reachableSet.size);
            const overlapPct = Math.round((intersect.size / minSize) * 100);
            overlaps.push({ e1, e2, intersectSize: intersect.size, overlapPct });
          }
        }
      }

      // Sort by overlap percentage descending
      overlaps.sort((a, b) => b.overlapPct - a.overlapPct);

      // Only show top 5 pairs
      const topOverlaps = overlaps.slice(0, 5);

      if (topOverlaps.length === 0) {
        container.innerHTML = '<div style="font-size: 0.8rem; color: var(--text-dim); text-align: center; padding: 1rem;">No overlaps detected.</div>';
        return;
      }

      topOverlaps.forEach(o => {
        const div = document.createElement('div');
        div.className = 'overlap-card';
        div.innerHTML = 
          '<div class="overlap-header">' +
            '<div class="overlap-names">' +
              '<span title="' + o.e1.route.path + '">' + o.e1.route.method + ' ' + o.e1.route.path + '</span>' +
              '<span style="color: var(--text-muted); font-size: 0.75rem;">and</span>' +
              '<span title="' + o.e2.route.path + '">' + o.e2.route.method + ' ' + o.e2.route.path + '</span>' +
            '</div>' +
            '<span class="overlap-badge">' + o.overlapPct + '%</span>' +
          '</div>' +
          '<div class="overlap-details">' +
            'They share <strong>' + o.intersectSize + '</strong> transitive function calls down the stack.' +
          '</div>';
        container.appendChild(div);
      });
    }

    // Step 5: Render Shared Functions Map
    function renderSharedFunctions(fullGroupEntries) {
      const container = document.getElementById('shared-funcs-container');
      container.innerHTML = '';

      // Count how many endpoints reach each function ID
      const funcCountMap = {};
      const funcCallerMap = {};

      fullGroupEntries.forEach(e => {
        e.reachableSet.forEach(funcID => {
          // Exclude the entry point itself to focus on shared library/database functions
          if (funcID !== e.id) {
            funcCountMap[funcID] = (funcCountMap[funcID] || 0) + 1;
            if (!funcCallerMap[funcID]) funcCallerMap[funcID] = [];
            funcCallerMap[funcID].push(e);
          }
        });
      });

      // Filter and get functions called by > 1 endpoints in the group
      const sharedFuncs = [];
      for (const [funcID, count] of Object.entries(funcCountMap)) {
        if (count > 1) {
          sharedFuncs.push({ funcID, count, callers: funcCallerMap[funcID] });
        }
      }

      // Sort by count descending
      sharedFuncs.sort((a, b) => b.count - a.count || a.funcID.localeCompare(b.funcID));

      if (sharedFuncs.length === 0) {
        container.innerHTML = '<div style="font-size: 0.8rem; color: var(--text-dim); text-align: center; padding: 1rem;">No shared transitive functions in this group.</div>';
        return;
      }

      sharedFuncs.forEach(sf => {
        const div = document.createElement('div');
        div.className = 'shared-func-item' + (selectedSharedFunc === sf.funcID ? ' active' : '');
        div.dataset.func = sf.funcID;
        div.innerHTML = 
          '<div class="shared-func-header">' +
            '<span style="word-break: break-all; color: var(--text-main); font-weight: 500;">' + sf.funcID + '</span>' +
            '<span class="shared-func-count">' + sf.count + ' callers</span>' +
          '</div>' +
          '<div style="font-size: 0.7rem; color: var(--text-muted); margin-top: 0.25rem;">' +
            'Reachable from: ' + sf.callers.map(c => c.route.method + ' ' + c.route.path).join(', ') +
          '</div>';
        div.onclick = (e) => {
          e.stopPropagation();
          toggleSharedFuncHighlight(sf.funcID, sf.callers.map(c => c.id));
        };
        container.appendChild(div);
      });
    }

    // Toggle endpoint card highlights when clicking a shared function
    function toggleSharedFuncHighlight(funcID, callerIDs) {
      const items = document.querySelectorAll('.shared-func-item');
      const cards = document.querySelectorAll('.endpoint-card');

      if (selectedSharedFunc === funcID) {
        selectedSharedFunc = '';
        items.forEach(el => el.classList.remove('active'));
        cards.forEach(el => el.classList.remove('active'));
      } else {
        selectedSharedFunc = funcID;
        items.forEach(el => {
          if (el.dataset.func === funcID) el.classList.add('active');
          else el.classList.remove('active');
        });
        
        cards.forEach(el => {
          if (callerIDs.includes(el.dataset.id)) {
            el.classList.add('active');
          } else {
            el.classList.remove('active');
          }
        });
      }
    }

    // Step 6: Slide-over Drawer for Endpoint Details
    function openEndpointDetails(entryID) {
      activeEntryID = entryID;
      const entry = entryNodes[entryID];
      if (!entry) {
        closeDrawer();
        return;
      }

      document.getElementById('drawer-endpoint-title').innerText = entry.route.method + ' ' + entry.route.path;
      document.getElementById('drawer-endpoint-fqn').innerText = entry.id;
      document.getElementById('drawer-rolled-cc').innerText = entry.rolledComplexity;
      document.getElementById('drawer-nodes-count').innerText = entry.reachableCount;
      document.getElementById('drawer-base-cc').innerText = entry.complexity;

      // Render overlaps inside drawer
      const overlapsList = document.getElementById('drawer-overlaps-list');
      overlapsList.innerHTML = '';

      const siblings = groups[activeGroup].filter(s => s.id !== entryID);
      const overlaps = [];

      siblings.forEach(sib => {
        const intersect = new Set([...entry.reachableSet].filter(x => sib.reachableSet.has(x)));
        if (intersect.size > 0) {
          const minSize = Math.min(entry.reachableSet.size, sib.reachableSet.size);
          const pct = Math.round((intersect.size / minSize) * 100);
          overlaps.push({ sib, count: intersect.size, pct });
        }
      });

      overlaps.sort((a, b) => b.pct - a.pct);

      if (overlaps.length === 0) {
        overlapsList.innerHTML = '<div style="font-size: 0.75rem; color: var(--text-dim);">No overlapping siblings.</div>';
      } else {
        overlaps.forEach(o => {
          const item = document.createElement('div');
          item.className = 'overlap-card clickable';
          item.style.padding = '0.5rem 0.75rem';
          item.innerHTML = 
            '<div class="overlap-header" style="font-size: 0.75rem;">' +
              '<span style="font-weight: 600; color: white;">' + o.sib.route.method + ' ' + o.sib.route.path + '</span>' +
              '<span class="overlap-badge" style="font-size: 0.7rem;">' + o.pct + '%</span>' +
            '</div>' +
            '<div style="font-size: 0.7rem; color: var(--text-dim);">Shares ' + o.count + ' functions</div>';
          
          item.onclick = () => {
            const isActive = item.classList.toggle('active');
            
            // Unselect other overlap cards
            document.querySelectorAll('.overlap-card').forEach(el => {
              if (el !== item) el.classList.remove('active');
            });
            
            // Clear previous shared highlights
            document.querySelectorAll('.drawer-list-item').forEach(el => {
              el.classList.remove('highlight-shared');
            });
            
            if (isActive) {
              // Highlight all shared dependencies
              const sharedSet = o.sib.reachableSet;
              document.querySelectorAll('.drawer-list-item').forEach(el => {
                const spanEl = el.querySelector('span');
                if (spanEl) {
                  const depID = spanEl.innerText;
                  if (sharedSet.has(depID)) {
                    el.classList.add('highlight-shared');
                  }
                }
              });
            }
          };
          
          overlapsList.appendChild(item);
        });
      }

      // Separate dependencies into boundaries and functions
      const boundaryList = [];
      const functionList = [];

      entry.reachableSet.forEach(depID => {
        if (!graph[depID]) {
          boundaryList.push(depID);
        } else {
          functionList.push(depID);
        }
      });

      // Sort boundaries alphabetically
      boundaryList.sort((a, b) => a.localeCompare(b));

      // Sort functions by complexity descending
      functionList.sort((a, b) => {
        const ccA = graph[a] ? graph[a].cyclomatic_complexity : 0;
        const ccB = graph[b] ? graph[b].cyclomatic_complexity : 0;
        if (ccA !== ccB) {
          return ccB - ccA;
        }
        return a.localeCompare(b);
      });

      // Render boundaries list
      const boundariesUI = document.getElementById('drawer-boundaries-list');
      boundariesUI.innerHTML = '';
      if (boundaryList.length === 0) {
        boundariesUI.innerHTML = '<div style="font-size: 0.75rem; color: var(--text-dim); padding: 0.5rem 0;">No method call boundaries.</div>';
      } else {
        boundaryList.forEach(depID => {
          const li = document.createElement('li');
          li.className = 'drawer-list-item';
          li.innerHTML = 
            '<div style="display: flex; justify-content: space-between;">' +
              '<span>' + depID + '</span>' +
              '<span style="color: var(--text-muted); font-size: 0.7rem;">Boundary</span>' +
            '</div>';
          boundariesUI.appendChild(li);
        });
      }

      // Render dependencies list
      const depsList = document.getElementById('drawer-dependencies-list');
      depsList.innerHTML = '';
      functionList.forEach(depID => {
        const li = document.createElement('li');
        li.className = 'drawer-list-item';
        
        // Highlight if this is the entry point itself
        if (depID === entryID) {
          li.style.borderColor = 'var(--accent-primary)';
          li.style.borderLeft = '2px solid var(--accent-primary)';
          li.style.color = 'white';
        }
        
        const cc = graph[depID] ? graph[depID].cyclomatic_complexity : 0;
        li.innerHTML = 
          '<div style="display: flex; justify-content: space-between;">' +
            '<span>' + depID + '</span>' +
            '<span style="color: var(--accent-teal); font-weight: 500;">CC: ' + cc + '</span>' +
          '</div>';
        depsList.appendChild(li);
      });

      // Open drawer animations
      document.getElementById('drawer-overlay').classList.add('open');
      document.getElementById('details-drawer').classList.add('open');
    }

    function closeDrawer() {
      activeEntryID = '';
      document.getElementById('drawer-overlay').classList.remove('open');
      document.getElementById('details-drawer').classList.remove('open');
    }

    // Global Noise Filtering
    let activeFilters = [];

    function updateActiveFilters() {
      const val = document.getElementById('filter-input').value.trim();
      if (!val) {
        activeFilters = [];
        return;
      }
      activeFilters = val.split(',').map(p => {
        const pattern = p.trim();
        if (!pattern) return null;
        const escaped = pattern.replace(/[-\/\\^$+?.()|[\]{}]/g, '\\$&');
        const regexStr = '^' + escaped.replace(/\*/g, '.*') + '$';
        return new RegExp(regexStr, 'i');
      }).filter(Boolean);
    }

    function isFilteredOut(id) {
      if (activeFilters.length === 0) return false;
      const parts = id.split('/');
      const lastPart = parts[parts.length - 1];
      const simpleName = lastPart.includes('.') ? lastPart.split('.').slice(1).join('.') : lastPart;

      for (const rx of activeFilters) {
        if (rx.test(id) || rx.test(lastPart) || rx.test(simpleName)) {
          return true;
        }
      }
      return false;
    }

    function onFilterChanged() {
      updateActiveFilters();
      localStorage.setItem('lucent_noise_filters', document.getElementById('filter-input').value);
      initializeData();
      renderSidebar();
      if (activeGroup) {
        selectGroup(activeGroup);
      }
      if (activeEntryID) {
        openEndpointDetails(activeEntryID);
      }
    }

    // Step 7: Search Filtering Event Listeners
    document.getElementById('search-box').addEventListener('input', (e) => {
      searchQuery = e.target.value;
      renderGroupDashboard();
    });

    document.getElementById('filter-input').addEventListener('input', onFilterChanged);

    document.getElementById('btn-preset-helpers').onclick = () => {
      document.getElementById('filter-input').value = '*Response, *Error, *Helper';
      onFilterChanged();
    };

    document.getElementById('btn-preset-loggers').onclick = () => {
      document.getElementById('filter-input').value = 'log.*, logger.*, ctx.*, context.*';
      onFilterChanged();
    };

    document.getElementById('btn-clear-filters').onclick = () => {
      document.getElementById('filter-input').value = '';
      onFilterChanged();
    };

    // Initialize Dashboard on load
    const savedFilters = localStorage.getItem('lucent_noise_filters');
    if (savedFilters !== null) {
      document.getElementById('filter-input').value = savedFilters;
    }
    updateActiveFilters();
    initializeData();
    renderSidebar();
    if (activeGroup) selectGroup(activeGroup);
  </script>
</body>
</html>`
