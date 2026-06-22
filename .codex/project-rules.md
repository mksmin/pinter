# Project Rules

## Go Formatting Style

Use multiline formatting for function signatures and calls when it improves readability.

Function signatures:

```go
func New(
	prefix string,
) (
	string,
	error,
) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b[:]), nil
}
```

Function and method calls:

```go
return s.hosts.AddHistory(
	ctx,
	host,
	command,
	terminalApp,
	startedAt,
)
```

Rules:

- Put every function signature argument on its own line.
- Put every return value on its own line when a signature has return values.
- Put call arguments on separate lines when there are two or more arguments.
- Keep a single short call argument inline, for example `formatTime(at)` or `ids.New("hst")`.
- Put a single long or chained call argument on its own line when inline form hurts readability.
- Keep trailing commas in multiline argument and return lists.
- Apply this style to new Go code and touched Go code.
