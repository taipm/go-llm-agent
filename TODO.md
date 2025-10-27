# TODO List - Active Tasks

> **Note**: Completed tasks have been moved to [DONE.md](DONE.md)

## Current Status

**Project**: go-llm-agent  
**Version**: v0.3.0 (85% Complete - 28 Built-in Tools Ready)  
**Last Updated**: October 27, 2025

---

## 🎯 v0.3.0 Planning - Advanced Tools

### Phase 2: Vector Database & Data Tools (Target: Nov-Dec 2025)
- [ ] qdrant_connect - Connect to Qdrant vector DB
- [ ] qdrant_create_collection - Create vector collection
- [ ] qdrant_upsert - Insert/update vectors
- [ ] qdrant_search - Semantic vector search
- [ ] qdrant_delete - Delete vectors

**Data Processing Tools (3 tools)** - Priority: MEDIUM
- [ ] data_json - JSON parsing and manipulation
- [ ] data_csv - CSV read/write/transform
- [ ] data_xml - XML parsing

**Status**: MongoDB tools COMPLETED, Qdrant tools in research phase (see RESEARCH_NEW_TOOLS.md)

---

## 📋 Documentation Updates Needed

- ✅ Update README.md with Network & Gmail tools examples
- ✅ Update CHANGELOG.md with Network & Gmail tools
- ✅ Update DONE.md and TODO.md with current status
- [ ] Add Network & Gmail tools to BUILTIN_TOOLS_DESIGN.md
- [ ] Create Qdrant design document
- [ ] Add MongoDB connection pooling best practices doc
- [ ] Gmail OAuth2 setup video tutorial (optional)

---

## 🚀 Release Planning

### v0.2.0 Release - READY
**Status**: Ready to release  
**Features**:
- Core agent framework
- 3 LLM providers (Ollama, OpenAI, Gemini)
- Memory management
- 13 built-in tools (File, Web, DateTime, System)
- Comprehensive examples

### v0.3.0 Release - IN PROGRESS (85% Complete)
**Target**: December 2025  
**Features**:
- ✅ Math tools (2 tools - COMPLETED Oct 27)
- ✅ MongoDB tools (5 tools - COMPLETED Oct 27)
- ✅ Network tools (5 tools - COMPLETED Oct 27)
- ✅ Gmail tools (4 tools - COMPLETED Oct 27)
- [ ] Qdrant vector search (5 tools - Planned)
- [ ] Data processing tools (3 tools - Planned)
- Current: 28 tools (24 auto-loaded + 4 Gmail opt-in) | Target: 40+ built-in tools total

---

## 🔄 Ongoing Maintenance

### Testing
- ✅ Maintain 80%+ code coverage
- ✅ All CI/CD pipelines green
- ✅ MongoDB tools: 7 test functions passing
- [ ] Add MongoDB integration tests with testcontainers
- [ ] Add Qdrant integration tests

### Performance
- ✅ Math tools tested with professional libraries
- [ ] Benchmark MongoDB connection pooling
- [ ] Optimize stat calculations for large datasets (>10k elements)
- [ ] Add caching for repeated calculations

### Security
- ✅ Expression evaluation safety (whitelist approach)
- ✅ MongoDB empty filter prevention (delete safety)
- ✅ Connection pool limits (max 10 connections)
- ✅ Network tools: DNS server validation, SSL verification
- ✅ Gmail tools: OAuth2 credential protection, token caching
- [ ] MongoDB connection string sanitization
- [ ] Qdrant API key management
- [ ] Rate limiting for database operations

---

## 📝 Notes

- **Professional Libraries Used**:
  - govaluate v3.0.0 (4.3k stars) - Expression evaluation
  - gonum v0.16.0 (7.2k stars) - Statistical operations
  - mongo-driver v1.17.4 (Official MongoDB Go driver)
  - miekg/dns v1.1.68 (Professional DNS library)
  - go-ping/ping v1.2.0 (ICMP ping)
  - likexian/whois v1.15.6 + whois-parser v1.24.20 (WHOIS queries)
  - oschwald/geoip2-golang v1.13.0 (IP geolocation)
  - google.golang.org/api v0.253.0 (Official Google Gmail API)
- **Current Status**: 28 tools registered in builtin package
- **Tool Categories**: 8 categories (File, Web, DateTime, System, Math, Database, Network, Email)
- **Safety**: 19/28 tools are safe (68% read-only operations)
- **Auto-loaded**: 24 tools (File, Web, DateTime, System, Math, Database, Network)
- **Opt-in**: 4 Gmail tools (requires OAuth2 credentials setup)
- **Examples**: 9 comprehensive demos with real-world use cases
- **Next Focus**: Qdrant vector search tools for v0.3.0
