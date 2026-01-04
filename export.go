package slogcloudwatch

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"
)

type ExportConfig struct {
	BatchSize   int // must be >=1
	Interval    time.Duration
	Destination LogDestination
	OrderedLogs bool
}

type LogDestination struct {
	LogGroupName  string
	LogStreamName string
}

type ExportOption func(*ExportConfig)

func NewExportConfig(opts ...ExportOption) ExportConfig {
	cfg := ExportConfig{
		BatchSize: 5,
		Interval:  5 * time.Second,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func WithBatchSize(n int) ExportOption {
	return func(ec *ExportConfig) {
		if n <= 0 {
			panic("[slog-cloudwatch] batch size must be >=1")
		}
		ec.BatchSize = n
	}
}

func WithInterval(d time.Duration) ExportOption {
	return func(ec *ExportConfig) {
		ec.Interval = d
	}
}

func WithLogGroupName(name string) ExportOption {
	return func(ec *ExportConfig) {
		ec.Destination.LogGroupName = name
	}
}

func WithLogStreamName(name string) ExportOption {
	return func(ec *ExportConfig) {
		ec.Destination.LogStreamName = name
	}
}

func EnableOrderedLogs() ExportOption {
	return func(ec *ExportConfig) {
		ec.OrderedLogs = true
	}
}

type BatchExporter struct {
	Client CloudWatchClient
	Queue  []LogEvent
	Config ExportConfig
}

func NewBatchExporter(client CloudWatchClient, cfg ExportConfig) *BatchExporter {
	return &BatchExporter{
		Client: client,
		Config: cfg,
	}
}

func (be *BatchExporter) Run(ctx context.Context, rx <-chan LogEvent, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(be.Config.Interval)
	defer ticker.Stop()

	flush := func() {
		if len(be.Queue) == 0 {
			return
		}
		logs := be.TakeFromQueue()

		putLogsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := be.Client.PutLogs(putLogsCtx, be.Config.Destination, logs); err != nil {
			fmt.Fprintf(os.Stderr,
				"[slog-cloudwatch] Unable to put logs to cloudwatch: %v %+v\n",
				err, be.Config.Destination)
		}
		cancel()
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		case ev, ok := <-rx:
			if !ok {
				flush()
				return
			}

			be.Queue = append(be.Queue, ev)
			if len(be.Queue) >= be.Config.BatchSize {
				flush()
			}
		}
	}
}

func (be *BatchExporter) TakeFromQueue() []LogEvent {
	logs := be.Queue
	be.Queue = nil

	if be.Config.OrderedLogs {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Timestamp.Before(logs[j].Timestamp)
		})
	}

	return logs
}
