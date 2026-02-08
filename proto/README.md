# Protocol Buffers

This directory contains the protocol buffer definitions for all services.

## Structure

```
proto/
├── todo/v1/        # Todo service protobuf definitions
└── id/v1/          # ID service protobuf definitions (future)
```

## Generated Code

Generated Go code is placed in `libs/proto-stubs/` and can be imported:

```go
import todov1 "platform/libs/proto-stubs/todo/v1"
```

## Generating

Run `just generate` to regenerate all protobufs and sqlc code.

## Adding a New Service

1. Create `proto/<service>/v1/<service>.proto`
2. Update `Justfile` to include the new proto path
3. Run `just generate`
