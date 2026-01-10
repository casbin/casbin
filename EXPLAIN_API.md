# Explain API Documentation

The Explain API provides AI-generated natural language explanations for authorization decisions made by Casbin's Enforce API.

## Overview

The Explain API uses an OpenAI-compatible API to generate human-readable explanations of why an authorization request was allowed or denied. This helps developers and administrators understand the access control logic and debug permission issues.

## Features

- **Uses only Go standard libraries** - No external dependencies beyond what Casbin already uses
- **OpenAI-compatible** - Works with OpenAI, Azure OpenAI, or any compatible API endpoint
- **Comprehensive context** - Sends model configuration, policies, request details, and enforcement result to the AI
- **Configurable** - Supports custom endpoints, models, timeouts, and API keys

## Quick Start

### 1. Configure the Explain API

```go
import (
    "time"
    "github.com/casbin/casbin/v3"
)

// Create enforcer
e, _ := casbin.NewEnforcer("model.conf", "policy.csv")

// Configure Explain API
e.SetExplainConfig(casbin.ExplainConfig{
    Endpoint: "https://api.openai.com/v1/chat/completions",
    APIKey:   "your-openai-api-key",
    Model:    "gpt-3.5-turbo", // or "gpt-4" for better explanations
    Timeout:  30 * time.Second,
})
```

### 2. Get Explanations

```go
// Check authorization
allowed, _ := e.Enforce("alice", "data1", "read")
fmt.Printf("Access allowed: %v\n", allowed)

// Get AI explanation
explanation, err := e.Explain("alice", "data1", "read")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Explanation:", explanation)
```

## Configuration Options

### ExplainConfig struct

```go
type ExplainConfig struct {
    // Endpoint is the API endpoint (required)
    // Examples:
    //   - OpenAI: "https://api.openai.com/v1/chat/completions"
    //   - Azure OpenAI: "https://<resource>.openai.azure.com/openai/deployments/<deployment>/chat/completions?api-version=2023-05-15"
    Endpoint string

    // APIKey is the authentication key for the API (required)
    APIKey string

    // Model is the model to use (required)
    // Examples: "gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"
    Model string

    // Timeout for API requests (optional, default: 30s)
    Timeout time.Duration
}
```

## Usage with Different Providers

### OpenAI

```go
e.SetExplainConfig(casbin.ExplainConfig{
    Endpoint: "https://api.openai.com/v1/chat/completions",
    APIKey:   "sk-...",
    Model:    "gpt-3.5-turbo",
})
```

### Azure OpenAI

```go
e.SetExplainConfig(casbin.ExplainConfig{
    Endpoint: "https://my-resource.openai.azure.com/openai/deployments/my-deployment/chat/completions?api-version=2023-05-15",
    APIKey:   "your-azure-key",
    Model:    "gpt-35-turbo", // Note: Azure uses different model naming
})
```

### Compatible Local Models

Any server implementing the OpenAI chat completions API format will work:

```go
e.SetExplainConfig(casbin.ExplainConfig{
    Endpoint: "http://localhost:8000/v1/chat/completions",
    APIKey:   "not-needed-for-local",
    Model:    "local-model",
})
```

## Example Output

For a request like `e.Explain("alice", "data1", "read")` where alice is allowed to read data1:

```
The authorization request was allowed because there is a matching policy rule 
that grants alice read permission on data1. The policy rule "p, alice, data1, read" 
explicitly allows this combination of subject, object, and action. The matcher in 
the model checks if the request parameters (alice, data1, read) match any policy 
rule, and in this case, it finds an exact match. Therefore, the effect is to allow 
the request.
```

For a denied request:

```
The authorization request was denied because there is no policy rule that allows 
alice to write to data1. While there is a rule allowing alice to read data1, there 
is no corresponding rule for the write action. The access control model requires 
an exact match between the request and a policy rule for access to be granted.
```

## Error Handling

The Explain API can fail for several reasons:

```go
explanation, err := e.Explain("alice", "data1", "read")
if err != nil {
    // Common errors:
    // - Config not set: "explain config not set, use SetExplainConfig first"
    // - Enforcement error: "failed to enforce: ..."
    // - API error: "failed to get AI explanation: ..."
    log.Printf("Failed to get explanation: %v", err)
}
```

## Best Practices

1. **Set timeout appropriately** - AI API calls can be slow, especially for complex policies
2. **Handle errors gracefully** - The Explain API is optional and should not block your main authorization flow
3. **Use for debugging** - Explain is most useful during development and troubleshooting
4. **Consider costs** - Each Explain call makes an API request to your AI provider, which may incur costs
5. **Cache explanations** - If you need to explain the same request multiple times, consider caching the results

## Limitations

- Requires external API access (OpenAI or compatible)
- Adds latency to authorization checks (use asynchronously for production)
- Explanation quality depends on the AI model used
- API costs may apply depending on your provider

## Implementation Details

The Explain API:
1. Calls `EnforceEx()` internally to get the enforcement result and matched rules
2. Builds a context string containing:
   - The authorization request (subject, object, action)
   - The enforcement result (allowed/denied)
   - Matched policy rules
   - Access control model configuration (matchers, effects)
   - All policy rules in the system
3. Sends this context to the configured AI API with a system prompt
4. Returns the AI-generated explanation

The implementation uses only Go standard libraries (`net/http`, `encoding/json`, `io`, etc.) to maintain Casbin's minimal dependency footprint.
