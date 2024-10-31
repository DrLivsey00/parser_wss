package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DrLivsey00/parser_wss/token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/<YOUR_SECRET_API_KEY>")
	if err != nil {
		panic(err)
	}

	contractAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") //contract address (USDC by default)

	tokenFilter, err := token.NewMainFilterer(contractAddress, client)

	sinc := make(chan *token.MainTransfer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchOpts := &bind.WatchOpts{
		Context: ctx,
	}

	sub, err := tokenFilter.WatchTransfer(watchOpts, sinc, nil, nil)
	defer sub.Unsubscribe()

	if err != nil {
		panic(err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Listening for transfer events...")

	go func() { //Listen for sub channell
		for event := range sinc {
			log.Printf("Tx hash: %s From: %s To: %s Value: %v",
				event.Raw.TxHash.Hex(),
				event.From.Hex(),
				event.To.Hex(),
				event.Tokens,
			)
		}
	}()
	select {
	case <-quit:
		fmt.Println("Shutting down gracefully...")
	case err := <-sub.Err():
		log.Printf("Subscription error: %v", err)
	}
	time.Sleep(2 * time.Second)
	fmt.Println("Shutdown complete.")
}
