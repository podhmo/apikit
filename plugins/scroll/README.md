# scroll

the plugin for scroll (pagination)

## how to use in code generation

```go
// pc is *plugins.PluginContext
// pkg is *tinypkg.Package
if err := pc.IncludePlugin(pkg, &scroll.Options{LatestID: ""}); err != nil {
    return err
}
```

## how to use in runtime

see [tests](./internal/scroll_test.go)
