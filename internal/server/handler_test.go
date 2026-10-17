package server

import (
	"testing"

	"github.com/jkqp/seechange/internal/buffer"
	"github.com/jkqp/seechange/internal/pb"
	"google.golang.org/protobuf/proto"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.Request
		want    *pb.Response
		prefill []string
		cap     int
	}{
		{name: "push request test success", req: &pb.Request{Op: &pb.Request_Push{Push: &pb.PushRequest{Value: "A"}}}, want: &pb.Response{Result: &pb.Response_Push{Push: &pb.PushResponse{Ok: true}}}, cap: 1},
		{name: "push request test fail", req: &pb.Request{Op: &pb.Request_Push{Push: &pb.PushRequest{Value: "A"}}}, want: &pb.Response{Result: &pb.Response_Push{Push: &pb.PushResponse{Ok: false}}}, prefill: []string{"A"}, cap: 1},
		{name: "pull request test success", req: &pb.Request{Op: &pb.Request_Pull{}}, want: &pb.Response{Result: &pb.Response_Pull{Pull: &pb.PullResponse{Value: "A", Ok: true}}}, prefill: []string{"A"}, cap: 1},
		{name: "pull request test fail", req: &pb.Request{Op: &pb.Request_Pull{}}, want: &pb.Response{Result: &pb.Response_Pull{Pull: &pb.PullResponse{}}}, cap: 1},
		{name: "size request test", req: &pb.Request{Op: &pb.Request_Size{}}, want: &pb.Response{Result: &pb.Response_Size{Size: &pb.SizeResponse{Size: 1}}}, prefill: []string{"A"}, cap: 1},
		{name: "empty request test", req: &pb.Request{}, want: &pb.Response{Result: &pb.Response_Err{Err: &pb.ErrorResponse{Message: "unknown or empty request"}}}, cap: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := buffer.New(tc.cap)
			for _, s := range tc.prefill {
				b.Push(s)
			}
			got := Handle(b, tc.req)
			if !proto.Equal(got, tc.want) {
				t.Errorf("test: %v failed, got: %v want:%v", tc.name, got, tc.want)
			}
		})
	}
}
