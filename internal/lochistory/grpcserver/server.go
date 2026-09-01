package grpcserver

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/service"
	pb "github.com/training/GOLANG_FOR_STUDENTS/pkg/proto"
)

// Server implements the LocationHistoryService gRPC interface.
type Server struct {
	pb.UnimplementedLocationHistoryServiceServer
	svc    service.Service
	logger *zap.Logger
}

// New creates a new gRPC server for the location history service.
func New(svc service.Service, logger *zap.Logger) *Server {
	return &Server{svc: svc, logger: logger}
}

// SaveLocation receives a location event from the location management service.
func (s *Server) SaveLocation(ctx context.Context, req *pb.SaveLocationRequest) (*pb.SaveLocationResponse, error) {
	ts, err := time.Parse(time.RFC3339, req.GetTimestamp())
	if err != nil {
		s.logger.Error("invalid timestamp in grpc request",
			zap.String("username", req.GetUsername()),
			zap.String("timestamp", req.GetTimestamp()),
			zap.Error(err),
		)
		return &pb.SaveLocationResponse{Success: false}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	if err := s.svc.SaveLocation(ctx, req.GetUsername(), req.GetLatitude(), req.GetLongitude(), ts); err != nil {
		s.logger.Error("failed to save location via grpc",
			zap.String("username", req.GetUsername()),
			zap.Error(err),
		)
		return &pb.SaveLocationResponse{Success: false}, err
	}

	s.logger.Info("location saved via grpc", zap.String("username", req.GetUsername()))
	return &pb.SaveLocationResponse{Success: true}, nil
}
