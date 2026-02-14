---
description: Run tests for the tempest CLI
command: go test ./... -race -count=1
---

# Test

Run all tests with race detection:
```bash
go test ./... -race -count=1
```

Run a specific package's tests:
```bash
go test ./internal/config -v
```

Run a specific test:
```bash
go test ./internal/config -run TestConfigLoad -v
```
