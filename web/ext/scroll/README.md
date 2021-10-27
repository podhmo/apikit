# scroll

the plugin for scroll (pagination)

## how to use in code generation

```go
// pc is *ext.PluginContext
// pkg is *tinypkg.Package
if err := pc.IncludePlugin(pkg, &scroll.Options{LatestID: ""}); err != nil {
    return err
}
```

## generated runtime

TODO: