package grep_usecase

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"dgrep/internal/config"
	"dgrep/internal/domain"
	"dgrep/internal/domain/errors"
)

type GrepService struct {
	cfg    *config.Config
	logger *zerolog.Logger
}

func NewGrepService(cfg *config.Config, logger *zerolog.Logger) domain.GrepUsecase {
	return &GrepService{cfg: cfg, logger: logger}
}

func (s *GrepService) ProcessLocalChunk(ctx context.Context, req domain.ChunkRequest) ([]domain.Match, error) {
	pattern := req.Pattern
	lines := req.Lines
	fixed := req.Fixed
	ignoreCase := req.IgnoreCase
	invert := req.Invert

	var matchFunc func(string) bool

	if fixed {
		if ignoreCase {
			lowerPattern := strings.ToLower(pattern)
			matchFunc = func(line string) bool {
				return strings.Contains(strings.ToLower(line), lowerPattern)
			}
		} else {
			matchFunc = func(line string) bool {
				return strings.Contains(line, pattern)
			}
		}
	} else {
		rePattern := pattern
		if ignoreCase {
			rePattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(rePattern)
		if err != nil {
			return nil, errors.ErrInvalidPattern
		}
		matchFunc = func(line string) bool {
			return re.MatchString(line)
		}
	}

	if invert {
		origMatch := matchFunc
		matchFunc = func(line string) bool {
			return !origMatch(line)
		}
	}

	numGoroutines := s.cfg.NumGoroutines
	subchunkSize := (len(lines) + numGoroutines - 1) / numGoroutines

	var wg sync.WaitGroup
	ch := make(chan []domain.Match, numGoroutines)
	errCh := make(chan error, 1)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		start := i * subchunkSize
		end := start + subchunkSize
		if end > len(lines) {
			end = len(lines)
		}
		subLines := lines[start:end]
		subStart := start

		go func(subLines []string, subStart int) {
			defer wg.Done()
			var localMatches []domain.Match
			for j, line := range subLines {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				default:
					if matchFunc(line) {
						localMatches = append(localMatches, domain.Match{RelLineNum: subStart + j, Text: line})
					}
				}
			}
			ch <- localMatches
		}(subLines, subStart)
	}

	wg.Wait()
	close(ch)
	close(errCh)

	if err, ok := <-errCh; ok && err != nil {
		return nil, err
	}

	var matches []domain.Match
	for local := range ch {
		matches = append(matches, local...)
	}

	return matches, nil
}

func (s *GrepService) DistributeAndCollect(ctx context.Context, repos []domain.GrepRepository, req domain.GrepRequest, chunkStarts []int) ([]domain.GlobalMatch, error) {
	N := len(repos)
	if N == 0 {
		return nil, errors.ErrNoServersProvided
	}

	lines := req.InputData
	chunkSize := (len(lines) + N - 1) / N
	var chunks [][]string
	for i := 0; i < N; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, lines[start:end])
		chunkStarts[i] = start
	}

	resultCh := make(chan domain.ChunkResponse, N)
	var wg sync.WaitGroup

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int, repo domain.GrepRepository) {
			defer wg.Done()
			chunkReq := domain.ChunkRequest{
				ChunkID:    i,
				Lines:      chunks[i],
				Pattern:    req.Pattern,
				Fixed:      req.Fixed,
				IgnoreCase: req.IgnoreCase,
				Invert:     req.Invert,
			}
			resp, err := repo.ProcessChunk(ctx, chunkReq)
			if err != nil {
				resultCh <- domain.ChunkResponse{ChunkID: i, Err: err}
				return
			}
			resultCh <- resp
		}(i, repos[i])
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(resultCh)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(s.cfg.Timeout):
		return nil, errors.ErrTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var allMatches []domain.GlobalMatch
	successCount := 0
	for res := range resultCh {
		if res.Err != nil {
			continue
		}
		successCount++
		chunkStart := chunkStarts[res.ChunkID]
		for _, m := range res.Matches {
			allMatches = append(allMatches, domain.GlobalMatch{GlobalLineNum: chunkStart + m.RelLineNum, Text: m.Text})
		}
	}

	quorum := N/2 + 1
	if successCount < quorum {
		return nil, fmt.Errorf("%w: %d/%d successful, need %d", errors.ErrQuorumNotReached, successCount, N, quorum)
	}

	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].GlobalLineNum < allMatches[j].GlobalLineNum
	})

	return allMatches, nil
}
