package srvu

import (
	"bufio"
	"context"
	"github.com/SimonSchneider/goslu/syncu"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type SSEResponseRecorder struct {
	*httptest.ResponseRecorder
	Body       *io.PipeWriter
	BodyReader *io.PipeReader
}

func NewSSEResponseRecorder() *SSEResponseRecorder {
	r := httptest.NewRecorder()
	pr, pw := io.Pipe()
	return &SSEResponseRecorder{ResponseRecorder: r, Body: pw, BodyReader: pr}
}

func (r *SSEResponseRecorder) Close() {
	r.Body.Close()
	r.BodyReader.Close()
}

func (r *SSEResponseRecorder) Write(b []byte) (int, error) {
	r.ResponseRecorder.Write(b)
	return r.Body.Write(b)
}

func (r *SSEResponseRecorder) Flush() {
	r.ResponseRecorder.Flush()
}

func TestSSEBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	broadcaster := syncu.NewBroadcaster[SSEEvent]()
	defer broadcaster.Close()
	broadcaster.Start()
	//handler := SSEHandler(broadcaster, 5*time.Second, 30*time.Second)
	handler := SSEHandler(broadcaster, 5*time.Second, 4*time.Millisecond)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	subscriberWg := &sync.WaitGroup{}
	subscriberWg.Add(2)
	broadcaster.OnSubscribe = func(c chan SSEEvent) {
		t.Logf("subscribed %p", c)
		subscriberWg.Done()
	}
	broadcaster.OnUnsubscribe = func(c chan SSEEvent) {
		t.Logf("unsubscribed %p", c)
	}
	ctx2, cancel2 := context.WithCancel(ctx)
	for i, ctx := range []context.Context{ctx, ctx2} {
		go func(i int, ctx context.Context) {
			defer wg.Done()
			w := NewSSEResponseRecorder()
			go func() {
				s := bufio.NewScanner(w.BodyReader)
				t.Logf("[%d]: started", i)
				for s.Scan() {
					t.Logf("[%d]: %s", i, s.Text())
				}
				t.Logf("[%d]: done", i)
			}()
			handler.ServeHTTP(w, httptest.NewRequestWithContext(ctx, "GET", "/", nil))
		}(i, ctx)
	}
	publishedHelloWorld := make(chan struct{})
	go func() {
		subscriberWg.Wait()
		broadcaster.Publish(SSEEvent{Data: "hello"})
		broadcaster.Publish(SSEEvent{Data: "world"})
		close(publishedHelloWorld)
		<-ctx2.Done()
		time.Sleep(10 * time.Millisecond)
		broadcaster.Publish(SSEEvent{Data: "goodbye"})
		broadcaster.Close()
	}()
	<-publishedHelloWorld
	time.Sleep(10 * time.Millisecond)
	cancel2()
	wg.Wait()
}
