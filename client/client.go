package client

import (
	proto "Handin5/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type clientInfo struct {
	client   proto.AuctionClient
	clientId int32
}

func (c *clientInfo) Bid(amount int32) {
	variable, err := c.client.Bid(context.Background(), &proto.Amount{Amount: amount})
	if err != nil {
		panic("PANIC!")
	}
	fmt.Println(variable)
}
func (c *clientInfo) Scanner() {
	reader := bufio.NewReader(os.Stdin)

	for true {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "exit":
			os.Exit(0)
		default:
			if input, err := strconv.Atoi(text); err == nil {
				c.Bid(int32(input))
			} else {
				fmt.Println("NOT A NUMBER!!!")
			}
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
