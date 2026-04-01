# vendel-sdk-go

Official Go SDK for the [Vendel](https://vendel.cc) SMS gateway API.

## Install

```bash
go get github.com/JimScope/vendel-sdk-go
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	vendel "github.com/JimScope/vendel-sdk-go"
)

func main() {
	client := vendel.NewClient("https://app.vendel.cc", "vk_your_api_key")

	// Send an SMS
	result, err := client.SendSMS(context.Background(), vendel.SendSMSRequest{
		Recipients: []string{"+1234567890"},
		Body:       "Hello from Vendel!",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.BatchID)

	// Check quota
	quota, err := client.GetQuota(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d/%d SMS used\n", quota.SMSSentThisMonth, quota.MaxSMSPerMonth)
}
```

## Webhook verification

```go
isValid := vendel.VerifyWebhookSignature(rawBody, signatureHeader, "your_webhook_secret")
```

## Requirements

- Go >= 1.21

## License

MIT
