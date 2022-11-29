package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	log "github.com/sirupsen/logrus"
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

func WithFileOutputWriter(filepath string) FileOutputWriter {
	return FileOutputWriter{filepath: filepath}
}

//var _ nba.OutputWriter = FileOutputWriter{}

type FileOutputWriter struct {
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
	r2Client cloudflare.Client,
	bucket string,
	objectKey string,
) R2OutputWriter {
	return R2OutputWriter{
		r2Client:  r2Client,
		bucket:    bucket,
		objectKey: objectKey,
	}
}

//var _ nba.OutputWriter = R2OutputWriter{}

type R2OutputWriter struct {
	r2Client  cloudflare.Client
	bucket    string
	objectKey string
}

func (r R2OutputWriter) Put(ctx context.Context, b []byte) error {
	if err := r.r2Client.CreateObject(ctx, r.bucket, r.objectKey, ContentTypeJSON, bytes.NewReader(b)); err != nil {
		log.WithError(err).WithFields(log.Fields{"bucket": r.bucket, "object_key": r.objectKey}).Error("failed to write object to r2 bucket")
	}

	return nil
}
