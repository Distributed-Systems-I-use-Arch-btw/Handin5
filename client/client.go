package client

import (
	proto "Handin5/grpc"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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
	logger   *log.Logger
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
		c.logger.Printf("Client, %d, got the result, all other responses are ignored\n", c.clientId)

		if outcome.Isover {
			resultChan <- fmt.Sprintf("Auction is over! 🔨\nHighest bid was: %d by client: %d", outcome.Highestbid, outcome.Clientid)
		} else {
			resultChan <- fmt.Sprintf("Auction is still going...\nHighest bid is: %d by client: %d", outcome.Highestbid, outcome.Clientid)
		}
	}
}

func (c *clients) fetchResults() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan string, len(c.clients))
	var wg sync.WaitGroup

	for _, client := range c.clients {
		wg.Add(1)
		go client.Result(ctx, resultChan, &wg)

		client.logger.Printf("Client, %d, asked its server for the result\n", client.clientId)
	}

	select {
	case result := <-resultChan:
		fmt.Println(result)
		cancel()
	case <-time.After(5 * time.Second):
		for _, client := range c.clients {
			client.logger.Println("No servers responded within the timeout period, shutting down...")
		}

		fmt.Println("No servers responded within the timeout period.")
		os.Exit(0)
	}

	wg.Wait()
	close(resultChan)
}

func (c *clientInfo) Bid(amount string) {
	ack, err := c.client.Bid(context.Background(),
		&proto.BidPackage{
			Amount:   &proto.Amount{Amount: amount},
			Clientid: &proto.ClientId{Clientid: c.clientId}})
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
			for _, client := range c.clients {
				client.logger.Printf("Client, %s, shut down\n", client.clientId)
			}
			os.Exit(0)
		case "result":
			c.fetchResults()
		default:
			for _, client := range c.clients {
				go client.Bid(text)
				client.logger.Printf("Client, %d, sent the bid, %d, to its server\n", client.clientId, text)
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

	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

	for _, port := range ports {
		conn, err = grpc.NewClient(fmt.Sprintf(
			"localhost:%d", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		client = proto.NewAuctionClient(conn)

		clients.clients = append(clients.clients, &clientInfo{
			client: client,
			logger: logger,
		})
	}
	if len(clients.clients) == 0 {
		fmt.Println("No servers are available. Exiting.")
		os.Exit(1)
	}

	for i := 0; i < len(clients.clients); i++ {
		client := clients.clients[i]
		clientId, err := client.client.CreateClientIdentifier(context.Background(), &proto.Empty{})
		if err != nil {
			clients.clients = append(clients.clients[:i], clients.clients[i+1:]...)
			i--
			continue
		}
		client.clientId = clientId.Clientid

		logger.Printf("Client, %d, has connected to a server on port %d\n", clientId.Clientid, ports[i])
	}

	clients.Scanner()
}
