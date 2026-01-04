package slogcloudwatch

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type Dispatcher interface {
	io.Writer
	Dispatch(input LogEvent)
	Stop()
}

type LogEvent struct {
	Message   string
	Timestamp time.Time
}

type CloudWatchDispatcher struct {
	tx     chan<- LogEvent
	cancel context.CancelFunc
	done   chan struct{}
}

func NewCloudWatchDispatcher(ctx context.Context, client *cloudwatchlogs.Client, ec ExportConfig) Dispatcher {
	ch := make(chan LogEvent, 1024)

	ctx, cancel := context.WithCancel(ctx)

	exporter := NewBatchExporter(&CloudwatchClient{client: client}, ec)

	dispatcher := CloudWatchDispatcher{
		tx:     ch,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go exporter.Run(ctx, ch, dispatcher.done)

	return &dispatcher
}

func (c *CloudWatchDispatcher) Dispatch(input LogEvent) {
	c.tx <- input
}

func (c *CloudWatchDispatcher) Stop() {
	c.cancel()
	<-c.done
}

func (c *CloudWatchDispatcher) Write(p []byte) (int, error) {
	c.Dispatch(LogEvent{
		Message:   string(p),
		Timestamp: time.Now().UTC(),
	})

	return len(p), nil
}
