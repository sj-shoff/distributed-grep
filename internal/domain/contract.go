package domain

import "context"

type GrepRepository interface {
	ProcessChunk(ctx context.Context, req ChunkRequest) (ChunkResponse, error)
}

type GrepUsecase interface {
	ProcessLocalChunk(ctx context.Context, req ChunkRequest) ([]Match, error)
	DistributeAndCollect(ctx context.Context, repos []GrepRepository, req GrepRequest, chunkStarts []int) ([]GlobalMatch, error)
}
