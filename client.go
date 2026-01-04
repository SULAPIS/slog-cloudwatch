package slogcloudwatch

type CloudWatchClient interface {
	PutLogs(dest LogDestination, logs []LogEvent) error
}

type LogDestinationNotFoundError struct {
	Message string
}

type OtherPutLogsError struct {
	Err error
}

func (e LogDestinationNotFoundError) Error() string {
	return e.Message
}

func (e OtherPutLogsError) Error() string {
	return e.Err.Error()
}
