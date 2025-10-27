# TODO List - Active Tasks

> **Note**: Completed tasks have been moved to [DONE.md](DONE.md)

## Current Status

**Project**: go-llm-agent  
**Version**: v0.3.0 (Phase 1 Complete - Math Tools Added)  
**Last Updated**: October 27, 2025

---

## ✅ Built-in Tools Phase 1 - COMPLETED (20/20 tools, 100%)

**Status**: ✅ COMPLETED (Oct 27, 2025)  
**Achievement**: All 20 Phase 1 built-in tools implemented, tested, and integrated  
**Total Lines**: ~7,000+ lines (production + tests + examples + docs)

### Completed Tools Summary

**File Tools (4/4)** ✅
- ✅ file_read - Read file content with security
- ✅ file_list - List directory with pattern matching
- ✅ file_write - Write/append with backup
- ✅ file_delete - Safe deletion with protection

**Web Tools (3/3)** ✅
- ✅ web_fetch - HTTP GET with SSRF prevention
- ✅ web_post - HTTP POST (JSON/form)
- ✅ web_scrape - CSS selector extraction

**DateTime Tools (3/3)** ✅
- ✅ datetime_now - Current time with formats
- ✅ datetime_format - Format/timezone conversion
- ✅ datetime_calc - Date calculations

**System Tools (3/3)** ✅
- ✅ system_info - CPU, memory, disk, OS, network
- ✅ system_processes - List/filter/sort processes
- ✅ system_apps - List installed applications

**Math Tools (2/2)** ✅
- ✅ math_calculate - Safe expression evaluation with govaluate
- ✅ math_stats - Statistical analysis with gonum/stat

**Database Tools (5/5)** ✅ **NEW**
- ✅ mongodb_connect - Connection pooling (max 10)
- ✅ mongodb_find - Query with filtering/sorting/projection
- ✅ mongodb_insert - Insert one or many (max 100 batch)
- ✅ mongodb_update - UpdateOne/UpdateMany with operators
- ✅ mongodb_delete - DeleteOne/DeleteMany with safety checks

### Integration Status
- ✅ Builtin package: GetRegistry() one-line setup
- ✅ Examples: 7 complete demos (including math_tools & mongodb_tools)
- ✅ All tests passing (200+ test cases)
- ✅ Security features: path validation, SSRF prevention, expression safety, empty filter prevention
- ✅ Cross-platform: macOS, Linux, Windows support
- ✅ Professional libraries: govaluate, gonum/stat, MongoDB driver

**Recent Commits**:
- cc7b935: Math tools implementation (calculate & stats)
- 561fcd4: Math tools example with 10 practical demos
- a239c80: Documentation updates for Math tools
- a8ce766: MongoDB tools implementation (connect, find, insert, update, delete)

---

## 🎯 v0.3.0 Planning - Advanced Tools

### Phase 2: Vector Database & Data Tools (Target: Nov-Dec 2025)

**MongoDB Tools (5 tools)** - ✅ COMPLETED (Oct 27, 2025)
- ✅ mongodb_connect - Connect with connection pooling
- ✅ mongodb_find - Query with filtering/sorting
- ✅ mongodb_insert - Batch insert (max 100)
- ✅ mongodb_update - UpdateOne/UpdateMany
- ✅ mongodb_delete - DeleteOne/DeleteMany with safety

**Qdrant Tools (5 tools)** - Priority: HIGH (NEXT)
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

- [ ] Update README.md with Math & MongoDB tools examples
- [ ] Add Math & MongoDB tools to BUILTIN_TOOLS_DESIGN.md
- [ ] Create Qdrant design document
- [ ] Add MongoDB connection pooling best practices doc

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

### v0.3.0 Release - IN PROGRESS (65% Complete)
**Target**: December 2025  
**Features**:
- ✅ Math tools (2 tools - COMPLETED Oct 27)
- ✅ MongoDB tools (5 tools - COMPLETED Oct 27)
- [ ] Qdrant vector search (5 tools - Planned)
- [ ] Data processing tools (3 tools - Planned)
- Current: 20 tools | Target: 30+ built-in tools total

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
- [ ] MongoDB connection string sanitization
- [ ] Qdrant API key management
- [ ] Rate limiting for database operations

---

## 📝 Notes

- **Professional Libraries Used**:
  * govaluate v3.0.0 (4.3k stars) - Expression evaluation
  * gonum v0.16.0 (7.2k stars) - Statistical operations
  * mongo-driver v1.17.4 (Official MongoDB Go driver)
- **Current Status**: 20 tools registered in builtin package
- **Tool Categories**: 6 categories (File, Web, DateTime, System, Math, Database)
- **Safety**: 15/20 tools are safe (75% read-only operations)
- **Examples**: 7 comprehensive demos with real-world use cases
- **Next Focus**: Qdrant vector search tools for v0.3.0
