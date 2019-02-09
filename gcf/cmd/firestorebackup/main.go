package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bm-sms/nomos/gcf"
)

func main() {
	ctx := context.Background()
	m := gcf.PubSubMessage{}
	err := gcf.BackupFirestore(ctx, m)
	if err != nil {
		fmt.Printf("Error: %+v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
