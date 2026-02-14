---
description: Build the tempest CLI binary
command: go build -o tempest .
---

# Build

Compile the CLI:
```bash
go build -o tempest .
```

Cross-compile with ldflags:
```bash
go build -ldflags "-X github.com/chadmayfield/tempest-cli/cmd.version=1.0.0 -X github.com/chadmayfield/tempest-cli/cmd.commit=$(git rev-parse --short HEAD) -X github.com/chadmayfield/tempest-cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o tempest .
```
