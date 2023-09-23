package util

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
)

func NBASeasonStageNameMappings() map[int]string {
	return map[int]string{
		1: "pre",
		2: "regular",
		3: "allstar",
		4: "post",
		5: "playin",
	}
}

func NBAGameStatusNameMappings() map[int]string {
	return map[int]string{
		1: "scheduled",
		2: "started",
		3: "completed",
	}
}

func WithFileOutputWriter(logger *slog.Logger, filepath string) FileOutputWriter {
	return FileOutputWriter{logger: logger, filepath: filepath}
}

//var _ nba.OutputWriter = FileOutputWriter{}

type FileOutputWriter struct {
	logger   *slog.Logger
	filepath string
}

func (f FileOutputWriter) Put(ctx context.Context, b []byte) error {
	dir := filepath.Dir(f.filepath)

	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("error creating directories when trying to write to file: %w", err)
	}

	file, err := os.Create(f.filepath)
	if err != nil {
		return fmt.Errorf("error creating file to write to: %w", err)
	}

	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

func WithR2OutputWriter(
	logger *slog.Logger,
	r2Client cloudflare.Client,
	bucket string,
	objectKey string,
) R2OutputWriter {
	return R2OutputWriter{
		logger:    logger,
		r2Client:  r2Client,
		bucket:    bucket,
		objectKey: objectKey,
	}
}

//var _ nba.OutputWriter = R2OutputWriter{}

type R2OutputWriter struct {
	logger    *slog.Logger
	r2Client  cloudflare.Client
	bucket    string
	objectKey string
}

func (r R2OutputWriter) Put(ctx context.Context, b []byte) error {
	if err := r.r2Client.CreateObject(ctx, r.bucket, r.objectKey, ContentTypeJSON, bytes.NewReader(b)); err != nil {
		r.logger.Error("failed to write object to r2 bucket", slog.String("bucket", r.bucket), slog.String("object_key", r.objectKey), slog.Any("error", err))
	}

	return nil
}
