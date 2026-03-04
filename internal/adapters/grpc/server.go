// internal/adapters/grpc/grpc_adapter.go
// Адаптер для gRPC: сервер (слушает, обрабатывает запросы), репозиторий (клиент для вызова gRPC).

package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"

	"dgrep/internal/domain"
	proto "dgrep/internal/gen/go/dgrep"
)

type GRPCGrepServer struct {
	proto.UnimplementedGrepServiceServer
	proto.UnimplementedHealthServer
	service domain.GrepUsecase
	logger  *zerolog.Logger
}

func (t *GRPCGrepServer) ProcessChunk(ctx context.Context, args *proto.ChunkRequest) (*proto.ChunkResponse, error) {
	req := domain.ChunkRequest{
		ChunkID:    int(args.ChunkId),
		Lines:      args.Lines,
		Pattern:    args.Pattern,
		Fixed:      args.Options.Fixed,
		IgnoreCase: args.Options.IgnoreCase,
		Invert:     args.Options.Invert,
	}
	matches, err := t.service.ProcessLocalChunk(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var protoMatches []*proto.Match
	for _, m := range matches {
		protoMatches = append(protoMatches, &proto.Match{
			LineNum: int32(m.RelLineNum),
			Text:    m.Text,
		})
	}
	return &proto.ChunkResponse{ChunkId: args.ChunkId, Matches: protoMatches}, nil
}

func (t *GRPCGrepServer) SearchStream(req *proto.SearchRequest, stream proto.GrepService_SearchStreamServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (t *GRPCGrepServer) Check(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{Status: proto.HealthCheckResponse_SERVING}, nil
}

func (t *GRPCGrepServer) Watch(req *proto.HealthCheckRequest, stream proto.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

type GRPCAdapter struct {
	service domain.GrepUsecase
	logger  *zerolog.Logger
}

func NewGRPCAdapter(service domain.GrepUsecase, logger *zerolog.Logger) *GRPCAdapter {
	return &GRPCAdapter{service: service, logger: logger}
}

func (a *GRPCAdapter) StartServer(ctx context.Context, addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Listen error")
	}
	defer l.Close()
	a.logger.Info().Msgf("Server listening on %s", addr)

	grpcServer := grpc.NewServer()
	proto.RegisterGrepServiceServer(grpcServer, &GRPCGrepServer{service: a.service, logger: a.logger})
	proto.RegisterHealthServer(grpcServer, &GRPCGrepServer{service: a.service, logger: a.logger})

	go func() {
		<-ctx.Done()
		a.logger.Info().Msg("Shutting down server")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(l); err != nil && err != grpc.ErrServerStopped {
		a.logger.Error().Err(err).Msg("gRPC serve error")
	}
}

type GRPCRepository struct {
	client proto.GrepServiceClient
}

func NewGRPCRepository(addr string) (domain.GrepRepository, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := proto.NewGrepServiceClient(conn)
	return &GRPCRepository{client: client}, nil
}

func (r *GRPCRepository) ProcessChunk(ctx context.Context, req domain.ChunkRequest) (domain.ChunkResponse, error) {
	protoReq := &proto.ChunkRequest{
		ChunkId: int32(req.ChunkID),
		Lines:   req.Lines,
		Pattern: req.Pattern,
		Options: &proto.GrepOptions{
			Fixed:      req.Fixed,
			IgnoreCase: req.IgnoreCase,
			Invert:     req.Invert,
		},
	}
	protoResp, err := r.client.ProcessChunk(ctx, protoReq)
	if err != nil {
		return domain.ChunkResponse{Err: err}, err
	}
	var matches []domain.Match
	for _, pm := range protoResp.Matches {
		matches = append(matches, domain.Match{
			RelLineNum: int(pm.LineNum),
			Text:       pm.Text,
		})
	}
	return domain.ChunkResponse{ChunkID: int(protoResp.ChunkId), Matches: matches}, nil
}
