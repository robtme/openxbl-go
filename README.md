# OpenXBL Go SDK
Interface with OpenXBL's API using Go.

**Note: This is in an experimental state. PRs are always welcome!**

## Setup

1. Create an account with [OpenXBL](https://xbl.io/)
2. Verify your email
3. Create an API key

## Login and Retrieve Account Info

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wolveix/openxbl-go"
)

func main() {
	client := openxbl.NewClient("your-api-token-here", time.Second*3)

	account, err := client.GetAccount(context.Background())
	if err != nil {
		log.Fatalf("Failed to query account info: %v", err)
	}

	fmt.Println(account)
}
```

## Roadmap

- [ ] Implement testing for all functions
- [ ] Support all API endpoints