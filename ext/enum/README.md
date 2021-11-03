# enum

the plugin for enum

## how to use in code generation

```go
// pc is *ext.PluginContext
// pkg is *tinypkg.Package

enumSet := enum.EnumSet{Name "Rank", Enums:[]enum.Enum{{Name: "Gold"}, {Name: "Silver"}, {Name: "Bronze"}}}
// or enumSet := enum.StringEnums("Rank", "Gold", "Silver", "Bronze")
if err := pc.IncludePlugin(pkg, &enum.Options{EnumSet: enumSet}); err != nil {
    return err
}
```

## generated code

an example is [here](./_examples/)