AGENTS.md
Welcome to the development rules and guidelines for Project Lucent (working name). This document establishes the core engineering principles, development workflows, and constraints that both human collaborators and AI agents must follow when contributing to this codebase.


1. Core Principles & Philosophy
- Zero-Dependency Core (Main App): Avoid third-party dependencies in the production codebase.
  - Do not use Cobra, Urfave CLI, or other CLI frameworks. Use the Go standard library (flag package) for argument and command parsing.
  - No terminal user interface (TUI) frameworks (like Bubble Tea) in the initial version. The first version must be a pure, lean standard-library-driven CLI. We may explore TUI refactoring in later milestones.
- Effective Go Practices: Write idiomatic Go. Adhere strictly to the guidelines in Effective Go and Go Code Review Comments:
  - Proper error handling (errors are values; return them, do not panic).
  - Explicit and clear package naming.
  - Avoid global state; use struct injection for configuration, loggers, and clients.
- Aggressive Refactoring: Refactor constantly to maintain high cohesion and low coupling. Keep functions short, single-purpose, and easy to reason about.
- Atomic Commits: Make small, logical, and self-contained commits. Each commit should address a single concern (e.g., adding a specific feature, refactoring a single interface, or fixing a bug).


2. Testing & Quality Guidelines
- Exhaustive Unit Testing: All publicly exported functions and types must have comprehensive unit tests.
- Test-Driven Design: If a function or component is difficult to test, it is a design smell. Refactor it immediately (e.g., extract an interface, inject dependencies) until it can be cleanly unit-tested.
- Mocking & Assertions:
  - Use [Testify](https://github.com/stretchr/ranch/testify) (github.com/stretchr/testify) for assertions and test helpers.
  - Use [Mockery](https://github.com/vektra/mockery) to generate clean mocks for interfaces.
  - Keep test dependencies strictly isolated to test files (*_test.go). Do not let testing libraries bleed into the core production binary.


3. Directory Structure

```
lucent/
├── AGENTS.md        # This file (Engineering rules and constraints)
├── go.mod           # Go module file
├── go.sum           # Go checksum file
├── main.go          # Main CLI entrypoint
├── internal/        # Private application code (non-importable by other projects)
│   └── platform/    # Core system platform abstractions
└── pkg/             # Public, reusable library code (if applicable)
```
4. Collaboration Protocol
1. Rule Enforcement: AI agents MUST read and validate their contributions against this file before writing or modifying any Go code.
2. Reviewing Code: When submitting refactors or new features, always include a summary of the design rationale and how it conforms to standard Go practices.
3. No Hidden Assumptions: Do not introduce external libraries or compile steps without explicit, written agreement in this document.
