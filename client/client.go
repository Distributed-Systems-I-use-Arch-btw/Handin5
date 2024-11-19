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
)

type clientInfo struct {
	client   proto.AuctionClient
	clientId int32
}

func (c *clientInfo) Result() {
	outcome, _ := c.client.Results(context.Background(), &proto.Empty{})

	if outcome.Isover {
		fmt.Println("Action is over!🔨")
		fmt.Printf("Highest bid was: %d \n", outcome.Highestbid)
	} else {
		fmt.Println("Action is still going...")
		fmt.Printf("Highest bid is: %d \n", outcome.Highestbid)
	}
}

func (c *clientInfo) Bid(amount string) {

	ack, _ := c.client.Bid(context.Background(), &proto.Amount{Amount: amount})
	if ack.Ack == "Exception" {
		fmt.Println("Not valid input")
	}

	fmt.Println(ack)
}

func (c *clientInfo) Scanner() {
	reader := bufio.NewReader(os.Stdin)

	for true {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "exit":
			os.Exit(0)
		case "result":
			c.Result()
		default:
			c.Bid(text)
		}
	}
}

func Run() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := proto.NewAuctionClient(conn)

	cliId, err := client.CreateClientIdentifier(context.Background(), &proto.Empty{})
	cliInfo := &clientInfo{client: client, clientId: cliId.Clientid}

	cliInfo.Scanner()
}
