package logging

import (
	"fmt"
	"sync"
)

var registry = newRegistry()

func newRegistry() *loggerRegistry {
	return &loggerRegistry{
		loggers: make(map[string]*Logger),
	}
}

type loggerRegistry struct {
	loggers map[string]*Logger
	mu      sync.Mutex
}

func (r *loggerRegistry) addLogger(l *Logger) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loggers[l.title]; ok {
		return fmt.Errorf("logger with name %s already exists", l.title)
	}
	r.loggers[l.title] = l
	return nil
}

func (r *loggerRegistry) addLoggerFromConfig(cfg Config) (*Logger, error) {
	l := NewFromConfig(cfg)
	if err := r.addLogger(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (r *loggerRegistry) getLogger(name string) (*Logger, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.loggers[name]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("logger with name %s does not exists", name)
}

func (r *loggerRegistry) setLevelForLogger(name string, l level) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if logger, ok := r.loggers[name]; ok {
		logger.level = l
	}
	return fmt.Errorf("logger with name %s does not exists", name)
}

func (r *loggerRegistry) setLevelForAll(l level) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, logger := range r.loggers {
		logger.level = l
	}
}

func (r *loggerRegistry) resetLevels() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, logger := range r.loggers {
		logger.resetLevel()
	}
}
