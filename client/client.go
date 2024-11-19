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
	"time"
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

	outcome, err := c.client.Results(context.Background(), &proto.Empty{})
	if err != nil {
		return
	}

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
	case <-time.After(5 * time.Second):
		fmt.Println("No servers responded within the timeout period.")
		os.Exit(0)
	}

	wg.Wait()
}

func (c *clientInfo) Bid(amount string) {
	ack, err := c.client.Bid(context.Background(), &proto.Amount{Amount: amount})
	if err != nil {
		return
	}
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
		conn, err = grpc.NewClient(fmt.Sprintf(
			"localhost:%d", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		client = proto.NewAuctionClient(conn)

		clients.clients = append(clients.clients, &clientInfo{client: client})
	}
	if len(clients.clients) == 0 {
		fmt.Println("No servers are available. Exiting.")
		os.Exit(1)
	}

	//var clientId *proto.ClientId

	for i := 0; i < len(clients.clients); i++ {
		client := clients.clients[i]
		clientId, err := client.client.CreateClientIdentifier(context.Background(), &proto.Empty{})
		if err != nil {
			clients.clients = append(clients.clients[:i], clients.clients[i+1:]...)
			i--
			continue
		}
		client.clientId = clientId.Clientid
	}

	clients.Scanner()
}
