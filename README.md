# slog-cloudwatch

slog-cloudwatch is a custom slog dispatcher that sends your Go application's logs to AWS CloudWatch Logs.

## Usage

### With AWS SDK

```go
package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	slogcloudwatch "github.com/sulapis/slog-cloudwatch"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	cwClient := cloudwatchlogs.NewFromConfig(cfg)

	dispatcher := slogcloudwatch.NewCloudWatchDispatcher(
		ctx,
		cwClient,
		slogcloudwatch.NewExportConfig(
			slogcloudwatch.WithLogGroupName("slog-cloudwatch"),
			slogcloudwatch.WithLogStreamName("stream-1"),
		),
	)

	slog.SetDefault(slog.New(slog.NewTextHandler(dispatcher, nil)))

	slog.Info("start")
	slog.Info("hello world")
	slog.Info("another log")
	slog.Error("a error")
	slog.Info("end")

	// Graceful shutdown: when Stop() is called or the context is canceled,
	// the exporter ensures that all logs currently in memory are flushed
	// to CloudWatch before exiting.
	// This guarantees that even if the program exits before the next scheduled ticker interval,
	// no log entries are lost.
	dispatcher.Stop()
}

```

#### Chronological order

When aggregating logs from multiple places, messages can become unordered. This causes a `InvalidParameterException: Log events in a single PutLogEvents request must be in chronological order.` error from the CloudWatch client. To mediate this, you may enable the `Ordered Logs` feature by calling `EnableOrderedLogs()` in your `ExportConfig`. Take into consideration that this can possibly increase processing time significantly depending on the number of events in the batch. Your milage may vary!

There is some additional context in https://github.com/ymgyt/tracing-cloudwatch/issues/40

## Required Permissions

Currently, following AWS IAM Permissions required

- `logs:PutLogEvents`

## CloudWatch Log Groups and Streams

This package does not create a log group and log stream, so if the specified log group and log stream does not exist, it will raise an error.

## Retry and Timeout

We haven't implemented any custom retry logic or timeout settings within the crate. We assume that these configurations are handled through the SDK Client.

## License

This project is licensed under the [MIT license.](./LICENSE)

## Acknowledgements

This project was inspired by [ymgyt/tracing-cloudwatch](https://github.com/ymgyt/tracing-cloudwatch).
