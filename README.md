# seer

A minimal opinionated error wrapper

## Installation

You can install directly with `go get` by running:

```sh
go get go.trulyao.dev/seer
```

## Usage

```go
func run() error {
	var val int

	fmt.Println("Enter an odd number: ")

	_, err := fmt.Scan(&val)
	if err != nil {
		return seer.Wrap("collectInput", err)
	}

	if val <= 0 {
		return seer.New("validateInput", fmt.Sprintf("negative number or zero: %d", val))
	}

	return nil
}
```

## Configuration

- `SetDefaultMessage`: This is used to set the default message that will be used as the wrapped error user-friendly message if no custom message is provided. Defaults to `"an error occurred"`.
- `SetCollectStackTrace`: Sets the `collectStackTrace` flag, when this is enabled, it collects additional data (filename, caller, line number etc) using the `runtime` package, this may have (tiny) performance costs if you use seer a lot, you should investigate if that is fine for your use case, in most cases, it is. Defaults to `true`.
