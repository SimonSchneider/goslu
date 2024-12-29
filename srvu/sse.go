package srvu

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SSEEvent struct {
	ID    string
	Event string
	Data  string
}

type WriteFlusher interface {
	http.ResponseWriter
	http.Flusher
}

type SSEWriter struct {
	WriteFlusher
}

func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, bool) {
	flusher, ok := w.(WriteFlusher)
	if !ok {
		return nil, false
	}
	return &SSEWriter{WriteFlusher: flusher}, true
}

func (w *SSEWriter) WriteComment(comment string) error {
	if strings.Contains(comment, "\n") {
		return fmt.Errorf("comment contains newline character")
	}
	if _, err := w.Write([]byte(": " + comment + "\n\n")); err != nil {
		return fmt.Errorf("write comment: %w", err)
	}
	w.Flush()
	return nil
}

func (w *SSEWriter) WriteRetry(retry time.Duration) error {
	if _, err := w.Write([]byte("retry: " + strconv.FormatInt(retry.Milliseconds(), 10) + "\n\n")); err != nil {
		return fmt.Errorf("write retry: %w", err)
	}
	w.Flush()
	return nil
}

func (w *SSEWriter) WriteEvent(e SSEEvent) error {
	if strings.Contains(e.ID, "\n") || strings.Contains(e.Event, "\n") {
		return fmt.Errorf("ID or Event contains newline character")
	}
	if e.ID != "" {
		if _, err := w.Write([]byte("id: " + e.ID + "\n")); err != nil {
			return fmt.Errorf("write id: %w", err)
		}
	}
	if e.Event != "" {
		if _, err := w.Write([]byte("event: " + e.Event + "\n")); err != nil {
			return fmt.Errorf("write event: %w", err)
		}
	}
	for _, line := range strings.Split(e.Data, "\n") {
		if _, err := w.Write([]byte("data: " + line + "\n")); err != nil {
			return fmt.Errorf("write data: %w", err)
		}
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("writing ending event newline: %w", err)
	}
	w.Flush()
	return nil
}

type SSESubscriber interface {
	Subscribe() (chan SSEEvent, func())
}

func SSEHandler(subscribe SSESubscriber, retry time.Duration, keepAliveInterval time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew, ok := NewSSEWriter(w)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		ew.Header().Set("Content-Type", "text/event-stream")

		if err := ew.WriteRetry(retry); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		source, unsub := subscribe.Subscribe()
		defer unsub()
		if source == nil {
			http.Error(w, "Server closed", http.StatusInternalServerError)
			return
		}
		keepAlive := time.NewTicker(keepAliveInterval)
		defer keepAlive.Stop()

		for {
			select {
			case <-ctx.Done():
				GetLogger(ctx).Printf("Client disconnected")
				return
			case <-keepAlive.C:
				if err := ew.WriteComment("keep-alive"); err != nil {
					GetLogger(ctx).Printf("Keep-alive error: %s", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case event, alive := <-source:
				if !alive {
					GetLogger(ctx).Printf("Event source closed")
					return
				}
				if err := ew.WriteEvent(event); err != nil {
					GetLogger(ctx).Printf("Event write error: %s", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	})
}
