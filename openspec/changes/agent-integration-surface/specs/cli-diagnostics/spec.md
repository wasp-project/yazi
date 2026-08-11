## ADDED Requirements

### Requirement: CLI Fails Without Panicking

The CLI SHALL NOT terminate via panic on any expected failure — unreachable
server, malformed JSON payload or filter, missing memory id, unavailable
provider, or a violated budget. It SHALL print a single human-readable error
line to stderr and exit with a non-zero status.

#### Scenario: Server is not running

- **WHEN** a memory command is run while no server is listening
- **THEN** the CLI prints a connection error naming the address and exits
  non-zero, with no Go stack trace in the output

#### Scenario: Malformed payload

- **WHEN** a put command is given invalid JSON
- **THEN** the CLI prints a parse error identifying the problem and exits
  non-zero

#### Scenario: Unavailable profile

- **WHEN** a recall requests a profile whose providers are unavailable
- **THEN** the CLI prints the provider name and its unmet requirements and exits
  non-zero, without a stack trace

### Requirement: Machine-Readable Success Output

Successful memory commands SHALL write only their result payload to stdout, so
the CLI can be composed in shell pipelines and driven by the benchmark harness.

#### Scenario: Recall output is parseable JSON

- **WHEN** a recall command succeeds
- **THEN** stdout contains only the JSON result document and diagnostics go to
  stderr

#### Scenario: Exit status signals success

- **WHEN** a memory command completes successfully
- **THEN** the process exits with status zero
