package snitch

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
)

// ErrorDetails hold additional information as part of the ErrorContext.
type ErrorDetails map[string]interface{}

// NewErrorDetails creates a new ErrorDetails structure.
func NewErrorDetails() ErrorDetails {
	return make(ErrorDetails)
}

// ErrorContext encodes the context of an error for reporting to an ErrorReporter.
type ErrorContext struct {
	Error   string
	Details ErrorDetails
}

// ErrorReporter is the interface for error reporting services.
type ErrorReporter interface {
	Notify(ectx *ErrorContext)
}

// LogReporter implements ErrorReporter via the log package.
type LogReporter struct {
	StackTraceDepth uint
}

// Notify notifies errors via the LogReporter.
func (l LogReporter) Notify(ectx *ErrorContext) {
	log.Println("Error:", ectx.Error)

	stack := make([]uintptr, l.StackTraceDepth)
	length := runtime.Callers(2, stack)
	for _, pc := range stack[:length] {
		if f := runtime.FuncForPC(pc); f != nil {
			file, line := f.FileLine(pc)
			log.Printf("  @ %s in %s:%d", f.Name(), file, line)
		}
	}
}

// MultiplexingReporter implements ErrorReporter which can notify multiple backends.
type MultiplexingReporter struct {
	ErrorReporters []ErrorReporter
	Mutex          sync.RWMutex
}

// Notify notifies errors via the MultiplexingReporter.
func (mr *MultiplexingReporter) Notify(ectx *ErrorContext) {
	mr.Mutex.RLock()
	defer mr.Mutex.RUnlock()

	for _, r := range mr.ErrorReporters {
		r.Notify(ectx)
	}
}

// AddReporter adds an ErrorReporter to a MultiplexingReporter.
func (mr *MultiplexingReporter) AddReporter(r ErrorReporter) {
	mr.Mutex.Lock()
	defer mr.Mutex.Unlock()

	mr.ErrorReporters = append(mr.ErrorReporters, r)
}

// PanicMonitor reports panics via the given ErrorReporter.
func PanicMonitor(er ErrorReporter) {
	if err := recover(); err != nil {
		if er != nil {
			er.Notify(&ErrorContext{
				Error: fmt.Sprintf("panic: %s", err),
			})
		}

		panic(err)
	}
}

// PanicMonitorHandler reports panics while handling HTTP requests via the given
// ErrorReporter.
func PanicMonitorHandler(er ErrorReporter, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PanicMonitor(er)
		h.ServeHTTP(w, r)
	})
}
