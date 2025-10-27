# Provider Comparison & Compatibility Guide

This document provides a detailed comparison of the three supported LLM providers: **Ollama**, **OpenAI**, and **Gemini**. It helps you choose the right provider for your use case and understand provider-specific behaviors.

## Quick Comparison Table

| Feature | Ollama | OpenAI | Gemini |
|---------|--------|--------|--------|
| **Deployment** | Local (self-hosted) | Cloud (API) | Cloud (API) |
| **API Key Required** | ❌ No | ✅ Yes | ✅ Yes |
| **Internet Required** | ❌ No | ✅ Yes | ✅ Yes |
| **Privacy** | ✅ Excellent (local) | ⚠️ Standard (cloud) | ⚠️ Standard (cloud) |
| **Cost** | ✅ Free | 💲 Pay-per-use | 💲 Pay-per-use |
| **Latency** | ✅ Very Low (local) | ⚠️ Network dependent | ⚠️ Network dependent |
| **Chat API** | ✅ Supported | ✅ Supported | ✅ Supported |
| **Streaming** | ✅ Supported | ✅ Supported | ✅ Supported |
| **Tool Calling** | ⚠️ Model-dependent | ✅ Fully Supported | ✅ Fully Supported |
| **Conversation History** | ✅ Supported | ✅ Supported | ✅ Supported |
| **Token Metadata** | ❌ Limited | ✅ Full | ✅ Full |
| **Max Context** | Model-dependent | 128K+ (GPT-4o) | 1M+ (Gemini 2.0) |
| **Response Speed** | ✅ Fast | ✅ Fast | ✅ Very Fast |
| **Availability** | Requires local install | ✅ Global | ✅ Global |

## Detailed Provider Analysis

### Ollama

**Best For:**
- Local development and testing
- Privacy-sensitive applications
- Offline environments
- Cost-conscious projects
- Learning and experimentation

**Strengths:**
- ✅ Completely free to use
- ✅ Runs locally - no internet required
- ✅ No API keys or authentication needed
- ✅ Full data privacy (nothing leaves your machine)
- ✅ Very low latency (no network overhead)
- ✅ Wide range of open-source models

**Limitations:**
- ❌ Requires local installation and setup
- ❌ Limited by your hardware (CPU/GPU/RAM)
- ❌ Tool calling support depends on model
- ❌ Less reliable metadata (token counts)
- ❌ Smaller context windows for most models
- ❌ Quality varies by model size

**Recommended Models:**
- `gemma3:4b` - Fast, good quality, low memory
- `llama3.2:3b` - Compact, efficient
- `qwen2.5:7b` - Better quality, more memory
- `mistral:7b` - Good all-rounder

**Configuration:**
```go
provider, err := provider.New(provider.Config{
    Type:    provider.ProviderOllama,
    BaseURL: "http://localhost:11434",
    Model:   "gemma3:4b",
})
```

### OpenAI

**Best For:**
- Production applications
- Advanced reasoning tasks
- Reliable tool calling
- Multi-modal needs (vision, audio)
- Enterprise applications

**Strengths:**
- ✅ State-of-the-art model quality
- ✅ Excellent tool/function calling
- ✅ Full token usage metadata
- ✅ Very large context windows (128K+)
- ✅ Highly reliable and fast
- ✅ Regular model updates
- ✅ Rich ecosystem and documentation

**Limitations:**
- 💲 Pay-per-use pricing (can add up)
- ⚠️ Requires internet connection
- ⚠️ Data sent to OpenAI servers
- ⚠️ API key management required
- ⚠️ Rate limits apply

**Recommended Models:**
- `gpt-4o-mini` - Fast, affordable, great quality
- `gpt-4o` - Best quality, higher cost
- `gpt-3.5-turbo` - Legacy, very affordable

**Pricing (as of Oct 2025):**
- `gpt-4o-mini`: $0.15/1M input, $0.60/1M output
- `gpt-4o`: $2.50/1M input, $10.00/1M output

**Configuration:**
```go
provider, err := provider.New(provider.Config{
    Type:   provider.ProviderOpenAI,
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "gpt-4o-mini",
})
```

**Azure OpenAI:**
```go
provider, err := provider.New(provider.Config{
    Type:    provider.ProviderOpenAI,
    APIKey:  os.Getenv("AZURE_OPENAI_API_KEY"),
    BaseURL: "https://your-resource.openai.azure.com",
    Model:   "gpt-4o",
})
```

### Gemini

**Best For:**
- Very large context requirements
- Multi-modal applications
- Cost-conscious cloud deployments
- Fast response times
- Google Cloud integrations

**Strengths:**
- ✅ Massive context windows (1M+ tokens)
- ✅ Very fast response times
- ✅ Competitive pricing
- ✅ Excellent tool calling
- ✅ Good multi-modal support
- ✅ Free tier available
- ✅ Vertex AI integration for enterprise

**Limitations:**
- 💲 Pay-per-use (though competitive)
- ⚠️ Requires internet connection
- ⚠️ Data sent to Google servers
- ⚠️ API key management required
- ⚠️ Newer ecosystem vs OpenAI

**Recommended Models:**
- `gemini-2.5-flash` - Fast, efficient, affordable
- `gemini-2.0-pro` - Better quality, higher cost
- `gemini-1.5-flash` - Legacy, very affordable

**Pricing (as of Oct 2025):**
- `gemini-2.5-flash`: Free tier available
- `gemini-2.0-pro`: Competitive with GPT-4o

**Configuration:**
```go
provider, err := provider.New(provider.Config{
    Type:   provider.ProviderGemini,
    APIKey: os.Getenv("GEMINI_API_KEY"),
    Model:  "gemini-2.5-flash",
})
```

**Vertex AI:**
```go
provider, err := provider.New(provider.Config{
    Type:       provider.ProviderGemini,
    ProjectID:  "your-gcp-project",
    Location:   "us-central1",
    Model:      "gemini-2.5-flash",
})
```

## Feature Compatibility Matrix

### Chat API

All three providers support the standard Chat API with identical signatures:

```go
response, err := provider.Chat(ctx, messages, options)
```

**Behavior Differences:**
- **Ollama**: May return empty metadata (no token counts)
- **OpenAI**: Always returns detailed token usage
- **Gemini**: Returns detailed token usage

**Testing Results** (compatibility_test.go):
- ✅ All providers handle simple questions correctly
- ✅ All providers remember conversation history
- ✅ Response times: Ollama (fastest), OpenAI (fast), Gemini (fastest)

### Streaming API

All three providers support streaming with identical callback interface:

```go
err := provider.Stream(ctx, messages, options, func(chunk StreamChunk) error {
    fmt.Print(chunk.Content)
    return nil
})
```

**Behavior Differences:**
- **Ollama**: Chunks can be very granular (character-by-character)
- **OpenAI**: Chunks are word/phrase based
- **Gemini**: Chunks are word/phrase based

**Testing Results:**
- ✅ All providers stream correctly
- ✅ All providers maintain conversation context in streaming
- ✅ Error handling is consistent

### Tool/Function Calling

**OpenAI:** ✅ **Full Support**
- Robust tool calling for all models
- Returns structured ToolCall objects
- Reliable argument parsing

**Gemini:** ✅ **Full Support**
- Excellent tool calling
- Returns structured ToolCall objects
- Good argument parsing

**Ollama:** ⚠️ **Model-Dependent**
- Only some models support tools
- gemma3:4b does NOT support tools
- llama3.2:3b and newer may support
- Returns error if model doesn't support: `"does not support tools"`

**Testing Results:**
- ✅ OpenAI: Full tool calling works
- ✅ Gemini: Full tool calling works
- ⚠️ Ollama (gemma3:4b): Gracefully returns error

### Error Handling

All providers handle errors consistently:

**Empty Message:**
- **Ollama**: Handles gracefully (may return generic response)
- **OpenAI**: May return error or handle gracefully
- **Gemini**: May return error or handle gracefully

**Empty History:**
- All providers handle empty message arrays gracefully

**Network Errors:**
- **Ollama**: Connection refused if not running
- **OpenAI**: HTTP 401 if invalid API key, 429 if rate limited
- **Gemini**: HTTP 401 if invalid API key, 429 if rate limited

## Performance Comparison

Based on compatibility tests (local development machine):

### Response Times

**Simple Math (2+2):**
- Ollama (gemma3:4b): ~4.0s (first call), ~0.3s (subsequent)
- OpenAI (gpt-4o-mini): ~0.5-1.0s
- Gemini (gemini-2.5-flash): ~0.3-0.7s

**Capital City Question:**
- Ollama: ~0.3s
- OpenAI: ~0.5s
- Gemini: ~0.4s

**Streaming (Count 1-5):**
- Ollama: ~0.5s (11 chunks)
- OpenAI: ~0.6s (varies)
- Gemini: ~0.4s (varies)

### Throughput

- **Ollama**: Limited by local hardware (CPU/GPU)
- **OpenAI**: Subject to rate limits (tier-dependent)
- **Gemini**: Subject to rate limits (generous free tier)

## Provider Selection Guide

### Choose **Ollama** if you:

- Want to develop/test without cost
- Need complete data privacy
- Work in offline environments
- Have sufficient local hardware
- Don't need advanced tool calling
- Are learning or experimenting

### Choose **OpenAI** if you:

- Need best-in-class quality
- Require reliable tool calling
- Want extensive documentation
- Have budget for API calls
- Need enterprise support
- Want multi-modal capabilities

### Choose **Gemini** if you:

- Need huge context windows (>100K tokens)
- Want fast responses
- Prefer Google Cloud ecosystem
- Need competitive pricing
- Want free tier for development
- Are building on GCP/Vertex AI

## Migration Guide

### Switching Between Providers

The factory pattern makes switching providers **zero-code-change**:

```bash
# .env file
LLM_PROVIDER=ollama    # or openai, gemini
LLM_MODEL=gemma3:4b    # or gpt-4o-mini, gemini-2.5-flash

# For Ollama
OLLAMA_BASE_URL=http://localhost:11434

# For OpenAI
OPENAI_API_KEY=sk-...

# For Gemini
GEMINI_API_KEY=...
```

```go
// Your code (works with ANY provider!)
provider, err := provider.FromEnv()
if err != nil {
    log.Fatal(err)
}

response, err := provider.Chat(ctx, messages, nil)
```

### Testing With Multiple Providers

```go
func TestAcrossProviders(t *testing.T) {
    providers := []struct{
        name string
        provider types.LLMProvider
    }{
        {"ollama", createOllamaProvider()},
        {"openai", createOpenAIProvider()},
        {"gemini", createGeminiProvider()},
    }

    for _, p := range providers {
        t.Run(p.name, func(t *testing.T) {
            // Same test runs across all providers
            response, err := p.provider.Chat(ctx, messages, nil)
            // ... assertions
        })
    }
}
```

## Known Limitations

### Ollama

1. **No Tool Calling** (for most models)
   - Workaround: Use prompt engineering to simulate tool calling
   - Or: Switch to OpenAI/Gemini for tool-heavy applications

2. **Limited Token Metadata**
   - Workaround: Estimate tokens manually or use tiktoken
   - Impact: Cannot track exact usage for billing/optimization

3. **Hardware Constraints**
   - Workaround: Use smaller models or upgrade hardware
   - Impact: Slower responses or quality trade-offs

### OpenAI

1. **Cost Can Add Up**
   - Workaround: Use gpt-4o-mini instead of gpt-4o
   - Workaround: Implement caching for repeated queries
   - Monitor: Track token usage via metadata

2. **Rate Limits**
   - Workaround: Implement retry with exponential backoff
   - Workaround: Request higher rate limits from OpenAI
   - Monitor: Track API usage in OpenAI dashboard

### Gemini

1. **Newer Ecosystem**
   - Workaround: Refer to Google AI documentation
   - Impact: Fewer community examples vs OpenAI

2. **Vertex AI Setup**
   - Workaround: Use standard Gemini API for simpler projects
   - Impact: More complex GCP authentication for Vertex AI

## Best Practices

### 1. Use Environment-Based Configuration

```go
// ✅ GOOD: Flexible, works with all providers
provider, err := provider.FromEnv()

// ❌ BAD: Hard-coded to one provider
provider := ollama.New("http://localhost:11434", "gemma3:4b")
```

### 2. Handle Provider-Specific Features Gracefully

```go
// Try tool calling, fall back if not supported
options := &types.ChatOptions{Tools: tools}
response, err := provider.Chat(ctx, messages, options)
if err != nil && strings.Contains(err.Error(), "does not support tools") {
    // Fallback: use regular chat
    response, err = provider.Chat(ctx, messages, nil)
}
```

### 3. Log Provider Being Used

```go
providerType := os.Getenv("LLM_PROVIDER")
model := os.Getenv("LLM_MODEL")
log.Printf("Using provider: %s, model: %s", providerType, model)
```

### 4. Test With Multiple Providers

Run your test suite against all providers to ensure compatibility:

```bash
# Test with Ollama
LLM_PROVIDER=ollama go test ./...

# Test with OpenAI
LLM_PROVIDER=openai go test ./...

# Test with Gemini
LLM_PROVIDER=gemini go test ./...
```

### 5. Monitor Costs (Cloud Providers)

```go
if response.Metadata != nil {
    totalTokens := response.Metadata.TotalTokens
    // Log or track for cost monitoring
    log.Printf("Tokens used: %d", totalTokens)
}
```

## Troubleshooting

### Ollama Not Responding

```bash
# Check if Ollama is running
ollama list

# Start Ollama
ollama serve

# Pull required model
ollama pull gemma3:4b
```

### OpenAI Authentication Errors

```bash
# Verify API key
echo $OPENAI_API_KEY

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Gemini Authentication Errors

```bash
# Verify API key
echo $GEMINI_API_KEY

# Check API is enabled in Google Cloud Console
# Visit: https://console.cloud.google.com/apis/library
```

## Conclusion

All three providers offer **identical APIs** through this library, making it easy to:

- ✅ Develop locally with Ollama (free)
- ✅ Test with multiple providers
- ✅ Deploy to production with OpenAI or Gemini
- ✅ Switch providers without code changes

Choose based on your **specific needs**: cost, privacy, quality, or features.

For most projects:
- **Development**: Ollama (free, fast iteration)
- **Production**: OpenAI or Gemini (quality, reliability)

The factory pattern ensures you can **start with one and migrate to another** seamlessly.
