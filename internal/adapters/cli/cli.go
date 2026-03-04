package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/rs/zerolog"

	grpc_adapter "dgrep/internal/adapters/grpc"
	"dgrep/internal/config"
	"dgrep/internal/domain"
	"dgrep/internal/domain/errors"
)

type CLIAdapter struct {
	service domain.GrepUsecase
	logger  *zerolog.Logger
}

func NewCLIAdapter(service domain.GrepUsecase, logger *zerolog.Logger) *CLIAdapter {
	return &CLIAdapter{service: service, logger: logger}
}

func (a *CLIAdapter) RunClient(ctx context.Context, addrsStr string, pattern string, inputFile string, stdin io.Reader, opts config.GrepOptions) {
	addrs := strings.Split(addrsStr, ",")

	var input io.Reader
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			a.logger.Fatal().Err(err).Msg("Failed to open input file")
		}
		defer file.Close()
		input = file
	} else {
		input = stdin
	}

	var lines []string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Client operation cancelled")
			return
		default:
			lines = append(lines, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		a.logger.Fatal().Err(err).Msg("Reading input")
	}

	var repos []domain.GrepRepository
	for _, addr := range addrs {
		repo, err := grpc_adapter.NewGRPCRepository(addr)
		if err != nil {
			a.logger.Warn().Err(err).Msgf("Failed to connect to %s", addr)
			repos = append(repos, &FailedRepository{})
			continue
		}
		repos = append(repos, repo)
	}

	chunkStarts := make([]int, len(addrs))
	grepReq := domain.GrepRequest{
		Pattern:    pattern,
		InputData:  lines,
		Fixed:      opts.Fixed,
		IgnoreCase: opts.IgnoreCase,
		Invert:     opts.Invert,
	}
	allMatches, err := a.service.DistributeAndCollect(ctx, repos, grepReq, chunkStarts)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Failed to distribute and collect")
	}

	a.processOutput(lines, allMatches, opts)

	a.logger.Info().Msg("Processed with quorum")
}

func (a *CLIAdapter) processOutput(lines []string, matches []domain.GlobalMatch, opts config.GrepOptions) {
	if opts.Count {
		fmt.Println(len(matches))
		return
	}

	after := opts.After
	before := opts.Before
	if opts.Context > 0 {
		after = opts.Context
		before = opts.Context
	}
	hasContext := after > 0 || before > 0

	if !hasContext {
		for _, m := range matches {
			line := m.Text
			if opts.LineNum {
				line = fmt.Sprintf("%d:%s", m.GlobalLineNum+1, line)
			}
			fmt.Println(line)
		}
		return
	}

	matchLines := make(map[int]struct{})
	for _, m := range matches {
		matchLines[m.GlobalLineNum] = struct{}{}
	}
	var matchNums []int
	for num := range matchLines {
		matchNums = append(matchNums, num)
	}
	sort.Ints(matchNums)

	var groups [][]int
	var currentGroup []int
	for _, num := range matchNums {
		if len(currentGroup) == 0 || num <= currentGroup[len(currentGroup)-1]+after+before {
			currentGroup = append(currentGroup, num)
		} else {
			groups = append(groups, currentGroup)
			currentGroup = []int{num}
		}
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	for gi, group := range groups {
		if gi > 0 {
			fmt.Println("--")
		}
		minLine := group[0] - before
		if minLine < 0 {
			minLine = 0
		}
		maxLine := group[len(group)-1] + after
		if maxLine >= len(lines) {
			maxLine = len(lines) - 1
		}

		for i := minLine; i <= maxLine; i++ {
			isMatch := false
			for _, gn := range group {
				if i == gn {
					isMatch = true
					break
				}
			}
			prefix := ""
			if opts.LineNum {
				prefix = fmt.Sprintf("%d", i+1)
			}
			sep := ":"
			if !isMatch {
				sep = "-"
			}
			if opts.LineNum {
				prefix += sep
			}
			fmt.Printf("%s%s\n", prefix, lines[i])
		}
	}
}

type FailedRepository struct{}

func (f *FailedRepository) ProcessChunk(ctx context.Context, req domain.ChunkRequest) (domain.ChunkResponse, error) {
	return domain.ChunkResponse{Err: errors.ErrConnectionFailed}, errors.ErrConnectionFailed
}
