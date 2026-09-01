package grpcclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/training/GOLANG_FOR_STUDENTS/pkg/proto"
)

// Client is the interface for sending location events to the history service.
type Client interface {
	SaveLocation(ctx context.Context, username string, lat, lon float64, ts time.Time) error
	Close() error
}

type grpcClient struct {
	conn   *grpc.ClientConn
	client pb.LocationHistoryServiceClient
}

// New dials the history service at addr and returns a Client.
func New(addr string) (Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial history service at %s: %w", addr, err)
	}
	return &grpcClient{
		conn:   conn,
		client: pb.NewLocationHistoryServiceClient(conn),
	}, nil
}

func (c *grpcClient) SaveLocation(ctx context.Context, username string, lat, lon float64, ts time.Time) error {
	req := &pb.SaveLocationRequest{
		Username:  username,
		Latitude:  lat,
		Longitude: lon,
		Timestamp: ts.Format(time.RFC3339),
	}
	resp, err := c.client.SaveLocation(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc SaveLocation for %q: %w", username, err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("history service reported failure for %q", username)
	}
	return nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
