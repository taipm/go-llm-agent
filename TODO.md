# TODO List - Active Tasks

> **Note**: Completed tasks have been moved to [DONE.md](DONE.md)

## Current Status

**Project**: go-llm-agent  
**Version**: v0.3.0 (Phase 1 Complete - Math Tools Added)  
**Last Updated**: October 27, 2025

---

## ✅ Built-in Tools Phase 1 - COMPLETED (15/15 tools, 100%)

**Status**: ✅ COMPLETED (Oct 27, 2025)  
**Achievement**: All 15 Phase 1 built-in tools implemented, tested, and integrated  
**Total Lines**: ~5,900 lines (production + tests + examples + docs)

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

**Math Tools (2/2)** ✅ **NEW**
- ✅ math_calculate - Safe expression evaluation with govaluate
- ✅ math_stats - Statistical analysis with gonum/stat

### Integration Status
- ✅ Builtin package: GetRegistry() one-line setup
- ✅ Examples: 6 complete demos including math_tools
- ✅ All tests passing (200+ test cases)
- ✅ Security features: path validation, SSRF prevention, expression safety
- ✅ Cross-platform: macOS, Linux, Windows support

**Recent Commits**:
- cc7b935: Math tools implementation (calculate & stats)
- 561fcd4: Math tools example with 10 practical demos

---

## 🎯 v0.3.0 Planning - Advanced Tools

### Phase 2: Database & Data Tools (Target: Nov-Dec 2025)

**MongoDB Tools (5 tools)** - Priority: HIGH
- [ ] mongodb_connect - Connect to MongoDB instance
- [ ] mongodb_find - Query documents
- [ ] mongodb_insert - Insert documents
- [ ] mongodb_update - Update documents
- [ ] mongodb_delete - Delete documents

**Qdrant Tools (5 tools)** - Priority: HIGH
- [ ] qdrant_connect - Connect to Qdrant vector DB
- [ ] qdrant_create_collection - Create vector collection
- [ ] qdrant_upsert - Insert/update vectors
- [ ] qdrant_search - Semantic vector search
- [ ] qdrant_delete - Delete vectors

**Data Processing Tools (3 tools)** - Priority: MEDIUM
- [ ] data_json - JSON parsing and manipulation
- [ ] data_csv - CSV read/write/transform
- [ ] data_xml - XML parsing

**Status**: Research phase complete (see RESEARCH_NEW_TOOLS.md)

---

## 📋 Documentation Updates Needed

- [ ] Update README.md with Math tools examples
- [ ] Add Math tools to BUILTIN_TOOLS_DESIGN.md
- [ ] Update DONE.md with Phase 1 completion
- [ ] Create MongoDB & Qdrant design documents

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

### v0.3.0 Release - IN PROGRESS  
**Target**: December 2025  
**Features**:
- Math tools (COMPLETED)
- MongoDB tools (Planned)
- Qdrant vector search (Planned)
- Data processing tools (Planned)
- Target: 25-30 built-in tools total

---

## 🔄 Ongoing Maintenance

### Testing
- ✅ Maintain 80%+ code coverage
- ✅ All CI/CD pipelines green
- [ ] Add integration tests for new DB tools

### Performance
- [ ] Benchmark math calculations
- [ ] Optimize stat calculations for large datasets
- [ ] Add caching for repeated calculations

### Security
- ✅ Expression evaluation safety (whitelist approach)
- [ ] MongoDB connection string validation
- [ ] Qdrant API key management
- [ ] Rate limiting for database operations

---

## 📝 Notes

- Math tools use professional libraries:
  * govaluate (4.3k stars) - Expression evaluation
  * gonum/stat (7.2k stars) - Statistical operations
- All 15 tools registered in builtin package
- Examples demonstrate real-world use cases
- Next focus: Database tools for v0.3.0
