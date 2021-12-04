package log

var traceEnabled = false

var DefaultLogger Logger

func SetLogger(logger Logger) {
	DefaultLogger = logger
}

// Trace logs to TRACE log. Arguments are handled in the manner of fmt.Print.
/*func Trace(args ...interface{}) {
	if traceEnabled {
		DefaultLogger.Trace(args...)
	}
}

// Tracef logs to TRACE log. Arguments are handled in the manner of fmt.Printf.
func Tracef(format string, args ...interface{}) {
	if traceEnabled {
		DefaultLogger.Tracef(format, args...)
	}
}

// TraceContext logs to TRACE log. Arguments are handled in the manner of fmt.Print.
func TraceContext(ctx context.Context, args ...interface{}) {
	if !traceEnabled {
		return
	}

	switch l := codec.Message(ctx).Logger().(type) {
	case *ZapLogWrapper:
		// 保护 l 或者 l.l 不可为空
		if l == nil || l.l == nil {
			DefaultLogger.Trace(args...)
			return
		}
		l.l.Trace(args...)
	case Logger:
		l.Trace(args...)
	default:
		DefaultLogger.Trace(args...)
	}
}

// TraceContextf logs to TRACE log. Arguments are handled in the manner of fmt.Printf.
func TraceContextf(ctx context.Context, format string, args ...interface{}) {
	if !traceEnabled {
		return
	}

	switch l := codec.Message(ctx).Logger().(type) {
	case *ZapLogWrapper:
		// 保护 l 或者 l.l 不可为空
		if l == nil || l.l == nil {
			DefaultLogger.Tracef(format, args...)
			return
		}
		l.l.Tracef(format, args...)
	case Logger:
		l.Tracef(format, args...)
	default:
		DefaultLogger.Tracef(format, args...)
	}
}*/

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Print.
func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func Warn(args ...interface{}) {
	DefaultLogger.Warn(args...)
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func Error(args ...interface{}) {
	DefaultLogger.Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func Fatal(args ...interface{}) {
	DefaultLogger.Fatal(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}
