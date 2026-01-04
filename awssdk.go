package slogcloudwatch

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudwatchClient struct {
	client *cloudwatchlogs.Client
}

func (s *CloudwatchClient) PutLogs(dest LogDestination, logs []LogEvent) error {
	logEvents := make([]types.InputLogEvent, 0, len(logs))

	for _, l := range logs {
		ile, err := l.ToInputLogEvent()
		if err != nil {
			return OtherPutLogsError{Err: err}
		}
		logEvents = append(logEvents, ile)
	}

	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     logEvents,
		LogGroupName:  &dest.LogGroupName,
		LogStreamName: &dest.LogStreamName,
	}
	output, err := s.client.PutLogEvents(context.Background(), input)
	if err != nil {
		var rnfe *types.ResourceNotFoundException
		if errors.As(err, &rnfe) {
			return LogDestinationNotFoundError{Message: *rnfe.Message}
		}
		return err
	}
	if output.RejectedLogEventsInfo != nil {
		fmt.Fprintf(os.Stderr, "[slog-cloudwatch] Put logs rejected: %+v\n", output.RejectedLogEventsInfo)
	}

	return nil
}

func (l *LogEvent) ToInputLogEvent() (types.InputLogEvent, error) {
	if l.Message == "" {
		return types.InputLogEvent{}, errors.New("log message is empty")
	}

	ts := l.Timestamp.UnixMilli()
	return types.InputLogEvent{
		Timestamp: &ts,
		Message:   &l.Message,
	}, nil
}
