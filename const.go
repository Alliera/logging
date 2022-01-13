package logging

const (
	Date = 1 << iota
	Time
	Labels
	Caller
	ShortCaller
)

const (
	DEBUG   level = "DEBUG"
	INFO    level = "INFO"
	WARNING level = "WARNING"
	ERROR   level = "ERROR"
	FATAL   level = "FATAL"
)

const (
	sourceErr = "UNKNOWN_SOURCE_ERROR"
)
