## ADDED Requirements

### Requirement: Explicit Configuration Path Selection

The server SHALL resolve its configuration file in this order: a `--config`
command-line flag, then a `YAZI_CONFIG` environment variable, then
`./config/default.yml`. The resolved path SHALL be logged at startup.

#### Scenario: Flag selects the configuration file

- **WHEN** the server is started with `--config /etc/yazi/prod.yml`
- **THEN** it loads that file and logs the path it used

#### Scenario: Environment variable is used when no flag is given

- **WHEN** `YAZI_CONFIG` is set and no `--config` flag is passed
- **THEN** the server loads the file named by the environment variable

### Requirement: Zero-Config Startup

When no configuration file exists at the default path, the server SHALL start
successfully using built-in defaults and SHALL log that it is doing so. When a
configuration file is requested explicitly but cannot be read or parsed, the
server SHALL fail fast with a clear error instead of silently falling back.

#### Scenario: Fresh clone starts without a config file

- **WHEN** the server is started in a working directory containing no
  `config/default.yml`
- **THEN** it starts with built-in defaults and serves memory operations

#### Scenario: Explicitly requested config that is missing is fatal

- **WHEN** the server is started with `--config /nonexistent.yml`
- **THEN** it exits non-zero with an error naming the unreadable path and does
  not start serving

#### Scenario: Malformed config is fatal

- **WHEN** the resolved configuration file contains invalid YAML
- **THEN** the server exits non-zero with a parse error and does not start with
  partially applied settings

### Requirement: Example Configurations Are Tracked

The repository SHALL track ready-to-use example configurations for the local and
cloud deployment modes. Only the operator's active configuration file SHALL be
ignored by version control.

#### Scenario: Examples exist in a fresh clone

- **WHEN** the repository is cloned
- **THEN** the local-mode and cloud-mode example configuration files are present
  and can be copied to the default path unmodified

#### Scenario: Operator config stays out of version control

- **WHEN** an operator edits the active default configuration file
- **THEN** version control reports no change for it
