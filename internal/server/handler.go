// Package server turns protobuf requests into calls on the buffer.
package server

import (
	"github.com/jkqp/seechange/internal/buffer"
	"github.com/jkqp/seechange/internal/pb"
)

// Handle runs the operation in req against b and returns the response.
func Handle(b *buffer.Buffer, req *pb.Request) *pb.Response {
	switch op := req.Op.(type) {
	case *pb.Request_Push:
		ok := b.Push(op.Push.GetValue())
		return &pb.Response{
			Result: &pb.Response_Push{Push: &pb.PushResponse{Ok: ok}},
		}
	case *pb.Request_Pull:
		value, ok := b.Pull()
		return &pb.Response{
			Result: &pb.Response_Pull{Pull: &pb.PullResponse{Value: value, Ok: ok}},
		}
	case *pb.Request_Size:
		size := b.Size()
		return &pb.Response{
			Result: &pb.Response_Size{Size: &pb.SizeResponse{Size: int32(size)}},
		}
	default:
		return &pb.Response{
			Result: &pb.Response_Err{Err: &pb.ErrorResponse{Message: "unknown or empty request"}},
		}
	}
}
