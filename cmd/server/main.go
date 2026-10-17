package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-zeromq/zmq4"
	"github.com/jkqp/seechange/internal/buffer"
	"github.com/jkqp/seechange/internal/pb"
	"github.com/jkqp/seechange/internal/server"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server error: %v", err)
	}

}

func run() error {
	// Cancel ctx on Ctrl+C or SIGTERM (what docker sends on stop) so the
	// server shuts down cleanly instead of being killed mid-loop.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	socket := zmq4.NewRep(ctx)
	defer socket.Close()
	if err := socket.Listen("tcp://*:5555"); err != nil {
		return fmt.Errorf("listening: %w", err)
	}
	log.Println("server listening on tcp://*:5555")
	b := buffer.New(10)
	for {
		msg, err := socket.Recv()
		if err != nil {
			if ctx.Err() != nil {
				log.Println("server shutting down")
				return nil
			}
			return fmt.Errorf("receiving: %w", err)
		}
		req := pb.Request{}
		err = proto.Unmarshal(msg.Bytes(), &req)
		if err != nil {
			resp := &pb.Response{
				Result: &pb.Response_Err{Err: &pb.ErrorResponse{Message: "invalid request"}},
			}
			reply, err := proto.Marshal(resp)
			if err != nil {
				return fmt.Errorf("cannot marshall response: %w", err)
			}
			if err := socket.Send(zmq4.NewMsg(reply)); err != nil {
				return fmt.Errorf("sending reply: %w", err)
			}
			continue
		}

		resp := server.Handle(b, &req)
		reply, err := proto.Marshal(resp)
		if err != nil {
			return fmt.Errorf("cannot marshall response: %w", err)
		}
		if err := socket.Send(zmq4.NewMsg(reply)); err != nil {
			return fmt.Errorf("sending reply: %w", err)
		}
	}
}
