# enum

the plugin for enum

## how to use in code generation

```go
// pc is *ext.PluginContext
// pkg is *tinypkg.Package

members := []enum.Enum{{Name: "Gold"}, {Name: "Silver"}, {Name: "Bronze"}}
if err := pc.IncludePlugin(pkg, &enum.Options{Name: "Grade", Members: members}); err != nil {
    return err
}
```

## generated code

an example is [here](./_examples)