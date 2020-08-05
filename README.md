# Seatbelt

A library for building web apps that speak HTML.

## API Reference

### `seatbelt`

#### `func New(config seatbelt.Config) *seatbelt.App`

Returns a new instance of a Seatbelt app.

```go
app := seatbelt.New(seatbelt.Config{})
app.Serve()
```

### `context`

### Template Helpers

#### `flashes`

Returns all flash messages as a `map[string]string`.

```gohtml
{{ range $level, $message := flashes }}
  <div class="flash-{{ $level }}">
    <p>{{ $message }}</p>
  </div>
{{ end }}
```
