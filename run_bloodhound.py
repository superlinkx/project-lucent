#!/usr/bin/env python3
# Copyright 2026 Alyx Holms
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import os
import re
import subprocess
import sys

def main():
    # 1. Compile the lucent binary
    print("Compiling lucent tool...")
    try:
        subprocess.run(["go", "build", "-o", "lucent"], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error compiling lucent: {e}")
        sys.exit(1)

    # 2. Paths to Bloodhound route files
    bloodhound_dir = "/var/home/alyx/Projects/Bloodhound"
    route_files = [
        os.path.join(bloodhound_dir, "cmd/api/src/api/registration/v2.go"),
        os.path.join(bloodhound_dir, "cmd/api/src/api/registration/registration.go")
    ]

    print("Extracting route handlers...")
    handlers = set()
    route_map = {}

    def get_group(path):
        parts = [p for p in path.strip('/').split('/') if p]
        filtered = [p for p in parts if p not in ("api", "v2", "v1", "{version}")]
        if filtered:
            return filtered[0]
        return "root"

    for file_path in route_files:
        if not os.path.exists(file_path):
            print(f"Warning: route file {file_path} not found.")
            continue
        
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        for line in content.splitlines():
            if "routerInst." in line:
                # Extract route method and path
                method = None
                path = None
                m_method = re.search(r'routerInst\.(GET|POST|PUT|DELETE|PATCH|PathPrefix)', line)
                if m_method:
                    method = m_method.group(1)
                    if method == "PathPrefix":
                        method = "ANY"
                
                m_path = re.search(r'"(/api/[^"]+)"', line)
                if m_path:
                    path = m_path.group(1)

                handler_fqn = None

                # Extract resources.<Method>
                m = re.search(r'\bresources\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    if method_name not in ["DB", "Config", "Authenticator", "Authorizer", "DogTags", "GraphQuery", "GraphDB"]:
                        handler_fqn = f"github.com/specterops/bloodhound/cmd/api/src/api/v2.(Resources).{method_name}"

                # Extract managementResource.<Method>
                m = re.search(r'\bmanagementResource\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    handler_fqn = f"github.com/specterops/bloodhound/cmd/api/src/api/v2/auth.(ManagementResource).{method_name}"

                # Extract loginResource.<Method>
                m = re.search(r'\bloginResource\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    handler_fqn = f"github.com/specterops/bloodhound/cmd/api/src/api/v2/auth.(LoginResource).{method_name}"

                # Extract v2.<Func>
                m = re.search(r'\bv2\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    if method_name not in ["Resources", "CollectorTypePathParameterName", "CollectorReleaseTagPathParameterName", "FileUploadJobIdPathParameterName", "CustomNodeKindParameter"]:
                        handler_fqn = f"github.com/specterops/bloodhound/cmd/api/src/api/v2.{method_name}"

                # Extract openapi.<Func>
                m = re.search(r'\bopenapi\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    handler_fqn = f"github.com/specterops/bloodhound/packages/go/openapi.{method_name}"

                # Extract static.<Func>
                m = re.search(r'\bstatic\.([a-zA-Z0-9_]+)\b', line)
                if m:
                    method_name = m.group(1)
                    handler_fqn = f"github.com/specterops/bloodhound/cmd/api/src/api/static.{method_name}"

                if handler_fqn:
                    handlers.add(handler_fqn)
                    if method and path:
                        route_map[handler_fqn] = {
                            "method": method,
                            "path": path,
                            "group": get_group(path)
                        }

    handlers_list = sorted(list(handlers))
    print(f"Extracted {len(handlers_list)} unique handler entry points.")
    print(f"Mapped {len(route_map)} route handlers to paths.")

    # Write route map to json file
    route_map_path = "lucent_route_map.json"
    with open(route_map_path, 'w', encoding='utf-8') as f:
        json.dump(route_map, f, indent=2)

    entry_arg = ",".join(handlers_list)

    # 3. Run lucent against the Bloodhound codebase
    print(f"Running lucent analysis on {bloodhound_dir}...")
    
    # We output to the local workspace directory
    cmd = [
        "./lucent",
        "-dir", bloodhound_dir,
        "-entry", entry_arg,
        "-format", "all",
        "-routes", route_map_path
    ]
    
    try:
        subprocess.run(cmd, check=True)
        print("Analysis completed successfully!")
        print("Output files generated in current directory:")
        print(" - lucent_opengraph.json  (BloodHound OpenGraph)")
        print(" - lucent_diagram.c4      (LikeC4 DSL)")
    except subprocess.CalledProcessError as e:
        print(f"Error running analysis: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
