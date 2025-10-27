# Agent Inspector

A diagnostic tool to inspect what capabilities an agent has immediately after creation with `agent.New(llm)`.

## Purpose

This tool answers the question: **"What does the agent have when I just call `agent.New(llm)`?"**

It displays:
1. Configuration settings (temperature, tokens, etc.)
2. Reasoning capabilities (CoT, ReAct, Reflection)
3. Memory system (type, features)
4. Learning system (status, readiness)
5. Built-in tools (categorized list)
6. Agent self-assessment (if available)

## Usage

```bash
cd examples/inspect_agent
go run main.go
```

Or build and run:

```bash
go build
./inspect_agent
```

## Example Output

```
🔍 Agent Inspection - What's Inside agent.New(llm)?
======================================================================

✅ Agent created with: agent.New(llm)

📋 1. CONFIGURATION
----------------------------------------------------------------------
System Prompt:       You are a helpful AI assistant.
Temperature:         0.7
Max Tokens:          2000
Max Iterations:      10
Min Confidence:      70.0%
Reflection Enabled:  true
Learning Enabled:    true

🧠 2. REASONING CAPABILITIES
----------------------------------------------------------------------
Auto Reasoning:      true
CoT Available:       false (Chain-of-Thought - step-by-step reasoning)
ReAct Available:     false (Reason + Act with tools)
Reflection:          false (Self-review and validation)

💾 3. MEMORY SYSTEM
----------------------------------------------------------------------
Type:                vector
Supports Search:     true
Supports Vectors:    true
Message Count:       0
✓ VectorMemory: Semantic search enabled (Qdrant detected)

📚 4. LEARNING SYSTEM
----------------------------------------------------------------------
Learning Enabled:    true
Experience Store:    false
Tool Selector:       false
Conversation ID:     aaccf768-6261-4b99-9c98-f3dfad8bc03a
⚡ Learning enabled but waiting for first interaction

🔧 5. BUILT-IN TOOLS
----------------------------------------------------------------------
Total Tools:         25

Available Tools:

  File Operations (4):
    • file_read
    • file_write
    • file_delete
    • file_list

  Web Tools (3):
    • web_fetch
    • web_post
    • web_scrape

  Network Tools (5):
    • network_dns_lookup
    • network_ping
    • network_whois_lookup
    • network_ssl_cert_check
    • network_ip_info

  DateTime Tools (3):
    • datetime_now
    • datetime_format
    • datetime_calc

  System Tools (3):
    • system_info
    • system_processes
    • system_apps

  Math Tools (2):
    • math_calculate
    • math_stats

  MongoDB Tools (5):
    • mongodb_connect
    • mongodb_find
    • mongodb_insert
    • mongodb_update
    • mongodb_delete

======================================================================
📊 SUMMARY
======================================================================
✅ Reasoning Modes:  3 (CoT, ReAct, Reflection)
✅ Built-in Tools:   25 tools ready to use
✅ Memory:           vector
✅ Learning:         true
✅ Auto Reasoning:   true

🎯 READY TO USE: Just call ag.Chat(ctx, "your question")
   The agent will automatically:
   - Select best reasoning mode (CoT/ReAct)
   - Use appropriate tools
   - Self-reflect on answers
   - Learn from experience
======================================================================
```

## What This Shows

### 1. **Zero Configuration**
- Agent is fully functional immediately after `agent.New(llm)`
- No manual setup required for any feature

### 2. **Lazy Initialization**
- Reasoning engines show as `false` initially
- They will be created automatically on first use
- Experience Store activates on first Chat() call

### 3. **Smart Defaults**
- Temperature: 0.7 (balanced creativity)
- Max Tokens: 2000 (reasonable responses)
- Reflection: Enabled (quality assurance)
- Learning: Enabled (continuous improvement)

### 4. **Memory Auto-Detection**
- With Qdrant: VectorMemory (semantic search)
- Without Qdrant: BufferMemory (simple storage)
- Graceful fallback, no errors

### 5. **Production-Ready Tools**
- 25+ tools loaded by default
- Organized by category
- Security built-in

## Technical Details

### Tool Categories

The inspector groups tools into logical categories:
- **File Operations**: File system access (read, write, list, delete)
- **Web Tools**: HTTP operations (fetch, post, scrape)
- **Network Tools**: Network diagnostics (DNS, ping, whois, SSL, IP)
- **DateTime Tools**: Time operations (now, format, calculate)
- **System Tools**: OS information (info, processes, apps)
- **Math Tools**: Mathematical operations (calculate, statistics)
- **MongoDB Tools**: Database operations (CRUD)

### Lazy Loading

Components marked as `false` don't mean they're unavailable:
- **CoT/ReAct/Reflection**: Created on first Chat() call
- **Experience Store**: Activated after first interaction
- **Tool Selector**: Ready after collecting enough experiences

This saves memory and startup time.

## Use Cases

### 1. **Debugging**
Check what the agent has before reporting issues:
```bash
./inspect_agent > agent_status.txt
```

### 2. **Verification**
Confirm setup is correct (especially Qdrant connection):
```
Memory Type: vector ✓
```

### 3. **Documentation**
Generate current capability list for your project docs.

### 4. **Development**
Quick reference for available tools and features during coding.

## Notes

- **Reasoning engines** appear as `false` initially - this is normal (lazy initialization)
- **Experience Store** needs first Chat() call to activate
- **Gmail tools** may appear if OAuth is configured (usually disabled)
- **IP info tool** may be absent if GeoIP database not found

## See Also

- [AGENT_CAPABILITIES.md](../../AGENT_CAPABILITIES.md) - Complete feature documentation
- [zero_config_agent](../zero_config_agent/) - Minimal usage example
- [learning_agent](../learning_agent/) - Learning system demo
