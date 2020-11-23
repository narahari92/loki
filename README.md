# loki
Loki is a chaos testing tool for cloud native applications with easy interfaces to write plugins and add new systems.

# Build locally
Run below command. `loki` binary will get installed in `$GOPATH/bin` directory.

```
go install github.com/narahari92/loki/cmd/loki
```

# Architecture
Please refer to [architecture.md](https://github.com/narahari92/loki/blob/master/docs/architecture.md) for details on architecture of loki.

# Testing plugins
To test whether a plugin implements loki correctly, simulate your system with resources, create `lokitest.Plugin` with the custom plugin's `System`, `Destroyer` and `killer`; and create `lokitest.Configuration` with some identifiers and sample destroy section. Then run below code to validate that the test passes.

```
lokitest.ValidateAll(context.Context, *testing.T, *lokitest.Plugin, *lokitest.Configuration)
```