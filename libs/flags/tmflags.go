package flags

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
)

const (
	defaultLogLevelKey = "*"
)

// ParseLogLevel parses complex log level - comma-separated
// list of module:level pairs with an optional *:level pair (* means
// all other modules).
//
// Example:
//		ParseLogLevel("consensus:debug,mempool:debug,*:error", log.NewTMLogger(os.Stdout), "info")
func ParseLogLevel(lvl string, logger log.Logger, defaultLogLevelValue string) (log.Logger, error) {
	if lvl == "" {
		return nil, errors.New("empty log level")
	}

	l := lvl

	// prefix simple one word levels (e.g. "info") with "*"
	if !strings.Contains(l, ":") {
		l = defaultLogLevelKey + ":" + l
	}

	options := make([]Option, 0)

	isDefaultLogLevelSet := false
	var option Option
	var err error

	list := strings.Split(l, ",")
	for _, item := range list {
		moduleAndLevel := strings.Split(item, ":")

		if len(moduleAndLevel) != 2 {
			return nil, fmt.Errorf("expected list in a form of \"module:level\" pairs, given pair %s, list %s", item, list)
		}

		module := moduleAndLevel[0]
		level := moduleAndLevel[1]

		if module == defaultLogLevelKey {
			option, err = AllowLevel(level)
			if err != nil {
				return nil, fmt.Errorf("failed to parse default log level (pair %s, list %s): %w", item, l, err)
			}
			options = append(options, option)
			isDefaultLogLevelSet = true
		} else {
			switch level {
			case "debug":
				option = AllowDebugWith("module", module)
			case "info":
				option = AllowInfoWith("module", module)
			case "error":
				option = AllowErrorWith("module", module)
			case "none":
				option = AllowNoneWith("module", module)
			default:
				return nil,
					fmt.Errorf("expected either \"info\", \"debug\", \"error\" or \"none\" log level, given %s (pair %s, list %s)",
						level,
						item,
						list)
			}
			options = append(options, option)

		}
	}

	// if "*" is not provided, set default global level
	if !isDefaultLogLevelSet {
		option, err = AllowLevel(defaultLogLevelValue)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	return NewFilter(logger, options...), nil
}

type level byte

const (
	levelDebug level = 1 << iota
	levelInfo
	levelError
)

type filter struct {
	next             log.Logger
	allowed          level            // XOR'd levels for default case
	initiallyAllowed level            // XOR'd levels for initial case
	allowedKeyvals   map[keyval]level // When key-value match, use this level
}

type keyval struct {
	key   interface{}
	value interface{}
}

// NewFilter wraps next and implements filtering. See the commentary on the
// Option functions for a detailed description of how to configure levels. If
// no options are provided, all leveled log events created with Debug, Info or
// Error helper methods are squelched.
func NewFilter(next log.Logger, options ...Option) log.Logger {
	l := &filter{
		next:           next,
		allowedKeyvals: make(map[keyval]level),
	}
	for _, option := range options {
		option(l)
	}
	l.initiallyAllowed = l.allowed
	return l
}

func (l *filter) Info(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelInfo != 0
	if !levelAllowed {
		return
	}
	l.next.Info(msg, keyvals...)
}

func (l *filter) Debug(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelDebug != 0
	if !levelAllowed {
		return
	}
	l.next.Debug(msg, keyvals...)
}

func (l *filter) Error(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelError != 0
	if !levelAllowed {
		return
	}
	l.next.Error(msg, keyvals...)
}

// With implements Logger by constructing a new filter with a keyvals appended
// to the logger.
//
// If custom level was set for a keyval pair using one of the
// Allow*With methods, it is used as the logger's level.
//
// Examples:
//     logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
//		 logger.With("module", "crypto").Info("Hello") # produces "I... Hello module=crypto"
//
//     logger = log.NewFilter(logger, log.AllowError(),
//				log.AllowInfoWith("module", "crypto"),
// 				log.AllowNoneWith("user", "Sam"))
//		 logger.With("module", "crypto", "user", "Sam").Info("Hello") # returns nil
//
//     logger = log.NewFilter(logger,
// 				log.AllowError(),
// 				log.AllowInfoWith("module", "crypto"), log.AllowNoneWith("user", "Sam"))
//		 logger.With("user", "Sam").With("module", "crypto").Info("Hello") # produces "I... Hello module=crypto user=Sam"
func (l *filter) With(keyvals ...interface{}) log.Logger {
	keyInAllowedKeyvals := false

	for i := len(keyvals) - 2; i >= 0; i -= 2 {
		for kv, allowed := range l.allowedKeyvals {
			if keyvals[i] == kv.key {
				keyInAllowedKeyvals = true
				// Example:
				//		logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
				//		logger.With("module", "crypto")
				if keyvals[i+1] == kv.value {
					return &filter{
						next:             l.next.With(keyvals...),
						allowed:          allowed, // set the desired level
						allowedKeyvals:   l.allowedKeyvals,
						initiallyAllowed: l.initiallyAllowed,
					}
				}
			}
		}
	}

	// Example:
	//		logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
	//		logger.With("module", "main")
	if keyInAllowedKeyvals {
		return &filter{
			next:             l.next.With(keyvals...),
			allowed:          l.initiallyAllowed, // return back to initially allowed
			allowedKeyvals:   l.allowedKeyvals,
			initiallyAllowed: l.initiallyAllowed,
		}
	}

	return &filter{
		next:             l.next.With(keyvals...),
		allowed:          l.allowed, // simply continue with the current level
		allowedKeyvals:   l.allowedKeyvals,
		initiallyAllowed: l.initiallyAllowed,
	}
}

//--------------------------------------------------------------------------------

// Option sets a parameter for the filter.
type Option func(*filter)

// AllowLevel returns an option for the given level or error if no option exist
// for such level.
func AllowLevel(lvl string) (Option, error) {
	switch lvl {
	case "debug":
		return AllowDebug(), nil
	case "info":
		return AllowInfo(), nil
	case "error":
		return AllowError(), nil
	case "none":
		return AllowNone(), nil
	default:
		return nil, fmt.Errorf("expected either \"info\", \"debug\", \"error\" or \"none\" level, given %s", lvl)
	}
}

// AllowAll is an alias for AllowDebug.
func AllowAll() Option {
	return AllowDebug()
}

// AllowDebug allows error, info and debug level log events to pass.
func AllowDebug() Option {
	return allowed(levelError | levelInfo | levelDebug)
}

// AllowInfo allows error and info level log events to pass.
func AllowInfo() Option {
	return allowed(levelError | levelInfo)
}

// AllowError allows only error level log events to pass.
func AllowError() Option {
	return allowed(levelError)
}

// AllowNone allows no leveled log events to pass.
func AllowNone() Option {
	return allowed(0)
}

func allowed(allowed level) Option {
	return func(l *filter) { l.allowed = allowed }
}

// AllowDebugWith allows error, info and debug level log events to pass for a specific key value pair.
func AllowDebugWith(key interface{}, value interface{}) Option {
	return func(l *filter) { l.allowedKeyvals[keyval{key, value}] = levelError | levelInfo | levelDebug }
}

// AllowInfoWith allows error and info level log events to pass for a specific key value pair.
func AllowInfoWith(key interface{}, value interface{}) Option {
	return func(l *filter) { l.allowedKeyvals[keyval{key, value}] = levelError | levelInfo }
}

// AllowErrorWith allows only error level log events to pass for a specific key value pair.
func AllowErrorWith(key interface{}, value interface{}) Option {
	return func(l *filter) { l.allowedKeyvals[keyval{key, value}] = levelError }
}

// AllowNoneWith allows no leveled log events to pass for a specific key value pair.
func AllowNoneWith(key interface{}, value interface{}) Option {
	return func(l *filter) { l.allowedKeyvals[keyval{key, value}] = 0 }
}
