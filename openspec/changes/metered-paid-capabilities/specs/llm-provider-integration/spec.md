## ADDED Requirements

### Requirement: Dependency-Free LLM Client

The system SHALL include an HTTP/JSON client for OpenAI-compatible chat-completion
and embedding endpoints, configurable by base URL, model and optional API key,
implemented without adding a module dependency, so a locally hosted model is a
first-class target.

#### Scenario: Client works against a local endpoint

- **WHEN** the client is configured with a local OpenAI-compatible base URL
- **THEN** completion and embedding requests succeed without any hosted-vendor
  credentials

#### Scenario: Token usage is captured from the response

- **WHEN** a completion returns usage information
- **THEN** the reported input and output token counts are recorded for metering

#### Scenario: Endpoint is logged at startup

- **WHEN** an LLM-backed provider is enabled
- **THEN** the configured endpoint is logged so operators can see where memories
  will be sent

### Requirement: Real Extractor And Reranker Providers

The LLM extractor and LLM reranker SHALL be implemented against the client,
report their token usage, and be selectable through profiles.

#### Scenario: Extractor turns raw input into structured memories

- **WHEN** the LLM extractor processes a raw input item
- **THEN** it returns one or more structured memory items and reports the tokens
  it consumed

#### Scenario: Reranker reorders candidates

- **WHEN** the LLM reranker receives candidate hits for a query
- **THEN** it returns them reordered and reports the tokens it consumed

#### Scenario: Prompts are inspectable

- **WHEN** an operator requests the prompts used by the LLM providers
- **THEN** the effective extraction and rerank prompts are printed

#### Scenario: Prompts are overridable

- **WHEN** an operator configures a custom extraction prompt
- **THEN** that prompt is used instead of the default

### Requirement: Availability Reflects Real Configuration

An LLM-backed provider SHALL report itself unavailable unless its endpoint is
configured and reachable, and selecting an unavailable provider SHALL fail at
startup with a message naming what is missing.

#### Scenario: Unconfigured pro profile fails at startup

- **WHEN** the full-capability profile is selected without LLM configuration
- **THEN** the server exits at startup with an error naming the missing settings,
  rather than failing on the first request

#### Scenario: Configured pro profile starts

- **WHEN** the full-capability profile is selected with a reachable endpoint
  configured
- **THEN** the server starts and serves extraction and reranking

### Requirement: LLM Failures Degrade Rather Than Corrupt

When an LLM call fails or times out, the operation SHALL fall back to the
deterministic path where a correct result is still possible, and SHALL report the
fallback, rather than returning a partial or fabricated result.

#### Scenario: Rerank failure falls back to deterministic order

- **WHEN** the reranker call fails during recall
- **THEN** the deterministically ranked hits are returned and the fallback is
  reported

#### Scenario: Extraction failure does not lose the write

- **WHEN** the extractor call fails during ingest
- **THEN** the original item is stored unextracted and the failure is reported
