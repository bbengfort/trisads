# TRISA Directory Service

**Implements a simple gRPC directory service for TRISA.**

This is a prototype implementation of a gRPC directory service that can act as a standalone server for VASP lookup queries. This is not intended to be used for production, but rather as a proof-of-concept (PoC) for directory service registration, lookups, and searches.

## Generate Protocol Buffers

To regenerate the Go code from the protocol buffers:

```
$ go generate ./...
```
