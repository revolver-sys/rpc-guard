package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	client := NewClient([]string{
		"https://ethereum-rpc.publicnode.com",
		"https://rpc.ankr.com/eth",
		"https://cloudflare-eth.com",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.Request(ctx, "eth_blockNumber", []any{})
	if err != nil {
		fmt.Println("request failed:", err)

		debug, _ := json.MarshalIndent(client.DebugState(), "", "  ")
		fmt.Println(string(debug))
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))

	debug, _ := json.MarshalIndent(client.DebugState(), "", "  ")
	fmt.Println(string(debug))
}
