# API Gateway

Lightweight API Gateway written in Go.

## Example


### Creating a new instace.

```go
func main() {
	gw, err := gateway.New()

	if err != nil {
		fmt.Printf("gateway create err: %v\n", err)

		return 
	}

	gw.Start()
}
```

