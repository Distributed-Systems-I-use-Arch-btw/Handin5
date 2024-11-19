package client

import (
	proto "Handin5/grpc"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strings"
	"sync"
)

type clients struct {
	clients []*clientInfo
}
type clientInfo struct {
	client   proto.AuctionClient
	clientId int32
}

func (c *clientInfo) Result(ctx context.Context, resultChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	outcome, _ := c.client.Results(context.Background(), &proto.Empty{})

	select {
	case <-ctx.Done():
		return
	default:
		if outcome.Isover {
			resultChan <- fmt.Sprintf("Auction is over! 🔨\nHighest bid was: %d", outcome.Highestbid)
		} else {
			resultChan <- fmt.Sprintf("Auction is still going...\nHighest bid is: %d", outcome.Highestbid)
		}
	}
}

func (c *clients) fetchResults() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan string, 1)
	var wg sync.WaitGroup

	for _, client := range c.clients {
		wg.Add(1)
		go client.Result(ctx, resultChan, &wg)
	}

	select {
	case result := <-resultChan:
		fmt.Println(result)
		cancel()
	}
}

func (c *clientInfo) Bid(amount string) {

	ack, _ := c.client.Bid(context.Background(), &proto.Amount{Amount: amount})
	if ack.Ack == "Exception" {
		fmt.Println("Not valid input")
	}

	fmt.Println(ack)
}

func (c *clients) Scanner() {
	reader := bufio.NewReader(os.Stdin)

	for true {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "exit":
			os.Exit(0)
		case "result":
			c.fetchResults()
		default:
			for _, client := range c.clients {
				go client.Bid(text)
			}
		}
	}
}

func Run() {
	clients := &clients{clients: make([]*clientInfo, 0)}
	ports := []int{5050, 5051, 5052}

	var conn *grpc.ClientConn
	var err error
	var client proto.AuctionClient

	for _, port := range ports {
		conn, err = grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		client = proto.NewAuctionClient(conn)

		clients.clients = append(clients.clients, &clientInfo{client: client})
	}

	for _, client := range clients.clients {
		var clientId *proto.ClientId
		clientId, _ = client.client.CreateClientIdentifier(context.Background(), &proto.Empty{})
		client.clientId = clientId.Clientid
	}

	clients.Scanner()
}
