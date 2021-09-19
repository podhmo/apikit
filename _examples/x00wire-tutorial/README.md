# wire-tutorial

- [wire/_tutorial at main · google/wire](https://github.com/google/wire/tree/main/_tutorial)

dependencies

```
Event -> Greeter -> Message -> phrase (string)
```


in wire, generated code is here.

```go
// Injectors from wire.go:

func InitializeEvent(phrase string) (Event, error) {
	message := NewMessage(phrase)
	greeter := NewGreeter(message)
	event, err := NewEvent(greeter)
	if err != nil {
		return Event{}, err
	}
	return event, nil
}
```