# Multi-Provider Chat Example

This example demonstrates the **factory pattern** for creating LLM providers. It supports **3 providers** (Ollama, OpenAI, Gemini) and shows **6 different configuration methods**.

## Features

- 🎯 **Auto-detect provider** from environment variables
- 🔧 **Manual configuration** for each provider
- ☁️ **Cloud platform support**: Azure OpenAI, Vertex AI
- 💬 **Interactive chat** with conversation history

## Quick Start

### 1. Setup Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env to select your provider
```

### 2. Choose Your Provider

#### Option A: Ollama (Local)
```bash
# .env
LLM_PROVIDER=ollama
LLM_MODEL=gemma3:4b
OLLAMA_BASE_URL=http://localhost:11434  # default
```

#### Option B: OpenAI
```bash
# .env
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o
OPENAI_API_KEY=sk-xxx
```

#### Option C: Azure OpenAI
```bash
# .env
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://mycompany.openai.azure.com
```

#### Option D: Gemini
```bash
# .env
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.5-flash
GEMINI_API_KEY=xxx
```

#### Option E: Vertex AI
```bash
# .env
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.5-flash
GEMINI_PROJECT_ID=my-gcp-project
GEMINI_LOCATION=us-central1
```

### 3. Run

```bash
go run main.go
```

## Six Ways to Create a Provider

### 1. Auto-detect from Environment (Recommended)

```go
provider, err := provider.FromEnv()
```

This reads `LLM_PROVIDER` and auto-configures based on environment variables.

### 2. Manual Config - Ollama

```go
provider, err := provider.New(provider.Config{
    Type:    provider.ProviderOllama,
    BaseURL: "http://localhost:11434",
    Model:   "gemma3:4b",
})
```

### 3. Manual Config - OpenAI

```go
provider, err := provider.New(provider.Config{
    Type:   provider.ProviderOpenAI,
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "gpt-4o",
})
```

### 4. Manual Config - Gemini

```go
provider, err := provider.New(provider.Config{
    Type:   provider.ProviderGemini,
    APIKey: os.Getenv("GEMINI_API_KEY"),
    Model:  "gemini-2.5-flash",
})
```

### 5. Manual Config - Azure OpenAI

```go
provider, err := provider.New(provider.Config{
    Type:    provider.ProviderOpenAI,
    APIKey:  os.Getenv("OPENAI_API_KEY"),
    BaseURL: "https://mycompany.openai.azure.com",
    Model:   "gpt-4o",
})
```

### 6. Manual Config - Vertex AI

```go
provider, err := provider.New(provider.Config{
    Type:      provider.ProviderGemini,
    ProjectID: "my-gcp-project",
    Location:  "us-central1",
    Model:     "gemini-2.5-flash",
})
```

## Environment Variables

| Variable | Description | Required For |
|----------|-------------|--------------|
| `LLM_PROVIDER` | Provider type: `ollama`, `openai`, `gemini` | All |
| `LLM_MODEL` | Model name | All |
| `OLLAMA_BASE_URL` | Ollama server URL | Ollama |
| `OPENAI_API_KEY` | OpenAI API key | OpenAI, Azure OpenAI |
| `OPENAI_BASE_URL` | OpenAI base URL (for Azure) | Azure OpenAI |
| `GEMINI_API_KEY` | Gemini API key | Gemini API |
| `GEMINI_PROJECT_ID` | GCP project ID | Vertex AI |
| `GEMINI_LOCATION` | GCP region | Vertex AI |

## Example Session

```
🤖 Multi-Provider Chat Demo
============================

✅ Using provider: ollama (model: gemma3:4b)

💬 Interactive Chat Mode
Type your message and press Enter. Type 'quit' to exit.

You: What is 2+2?
Assistant: 2 + 2 = 4

You: Tell me a joke
Assistant: Why don't scientists trust atoms? Because they make up everything! 😊

You: quit
👋 Goodbye!
```

## Testing Different Providers

Switch between providers by changing `.env`:

```bash
# Test with Ollama
echo "LLM_PROVIDER=ollama" > .env
echo "LLM_MODEL=gemma3:4b" >> .env
go run main.go

# Test with OpenAI
echo "LLM_PROVIDER=openai" > .env
echo "LLM_MODEL=gpt-4o" >> .env
echo "OPENAI_API_KEY=sk-xxx" >> .env
go run main.go

# Test with Gemini
echo "LLM_PROVIDER=gemini" > .env
echo "LLM_MODEL=gemini-2.5-flash" >> .env
echo "GEMINI_API_KEY=xxx" >> .env
go run main.go
```

## Code Structure

- `main.go`: Main program
- `.env`: Environment variables (create from `.env.example`)
- `.env.example`: Template for environment configuration

## Next Steps

- Try different models for each provider
- Test Azure OpenAI and Vertex AI configurations
- Integrate factory pattern into your own applications
- Explore streaming and tool calling features
