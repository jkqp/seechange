package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-zeromq/zmq4"
	"github.com/jkqp/seechange/internal/pb"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("client error: %v", err)
	}
}
func run() error {
	ctx := context.Background()
	socket := zmq4.NewReq(ctx, zmq4.WithDialerRetry(time.Second))
	defer socket.Close()

	// SERVER_ADDR is set by docker compose; fall back to localhost for go run.
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "tcp://localhost:5555"
	}
	if err := socket.Dial(addr); err != nil {
		return fmt.Errorf("dialing: %w", err)
	}
	requests := []*pb.Request{
		pushRequest("A"),
		pushRequest("B"),
		pushRequest("C"),
		sizeRequest(),
		pullRequest(),
		sizeRequest(),
		pullRequest(),
		pullRequest(),
		pullRequest(),
	}

	for _, req := range requests {
		resp, err := send(socket, req)
		if err != nil {
			return err
		}
		fmt.Printf("%-16s -> %s\n", req.String(), resp.String())
	}
	return nil
}

func pushRequest(value string) *pb.Request {
	return &pb.Request{Op: &pb.Request_Push{Push: &pb.PushRequest{Value: value}}}
}

func pullRequest() *pb.Request {
	return &pb.Request{Op: &pb.Request_Pull{Pull: &pb.PullRequest{}}}
}

func sizeRequest() *pb.Request {
	return &pb.Request{Op: &pb.Request_Size{Size: &pb.SizeRequest{}}}
}

// send makes one round trip to the server: it marshals req, sends it,
// waits for the reply and unmarshals it into a Response.
func send(socket zmq4.Socket, req *pb.Request) (*pb.Response, error) {
	msg, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("cannot marshall request: %w", err)
	}
	if err := socket.Send(zmq4.NewMsg(msg)); err != nil {
		return nil, fmt.Errorf("sending: %w", err)
	}

	r, err := socket.Recv()
	if err != nil {
		return nil, fmt.Errorf("receiving: %w", err)
	}
	resp := &pb.Response{}
	if err := proto.Unmarshal(r.Bytes(), resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshall response: %w", err)
	}
	return resp, nil
}
