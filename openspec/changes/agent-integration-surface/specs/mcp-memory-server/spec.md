## ADDED Requirements

### Requirement: MCP Stdio Server

The binary SHALL provide a mode that speaks the Model Context Protocol over
JSON-RPC 2.0 on stdin/stdout, so an MCP-capable agent can attach Yazi through
configuration alone, without writing glue code.

#### Scenario: Agent initializes the server

- **WHEN** an MCP client sends an initialize request on stdin
- **THEN** the server responds with its protocol version and declared tool
  capability, and writes nothing else to stdout

#### Scenario: Logs never corrupt the protocol stream

- **WHEN** the server emits a log line while serving MCP
- **THEN** the line is written to stderr and stdout carries only JSON-RPC
  messages

### Requirement: Minimal Memory Tool Surface

The MCP server SHALL expose exactly four tools — remember, recall,
list_memories, and forget — because tool descriptions are prompt tokens the
agent pays on every turn. Each tool SHALL have a description short enough to be
carried cheaply and SHALL map onto the server-side memory service.

#### Scenario: Tool list is stable and minimal

- **WHEN** an MCP client lists available tools
- **THEN** it receives exactly the four memory tools with their input schemas

#### Scenario: Remember stores a structured record

- **WHEN** the agent calls remember with content and optional kind, scope,
  subject and tags
- **THEN** a basic memory is persisted and the tool returns its id

#### Scenario: Recall returns memories and their cost

- **WHEN** the agent calls recall with a query
- **THEN** the tool returns the matching memory texts and the metered usage and
  estimated cost of producing them

#### Scenario: Forget removes a memory

- **WHEN** the agent calls forget with a memory id
- **THEN** the record and its index entries are removed and the tool confirms
  the deletion

### Requirement: MCP Honors Tenancy And Budgets

MCP tool calls SHALL flow through the same tenant resolution, profile selection,
and budget enforcement as every other transport.

#### Scenario: Recall respects a context-token budget

- **WHEN** a recall budget limits context tokens and the matching memories
  exceed it
- **THEN** the tool returns a trimmed set of memories within the budget rather
  than the full set
