package writer

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
)

type Writer struct {
	w http.ResponseWriter

	status int
	buffer bytes.Buffer
}

func Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		instance := &Writer{w: w, status: 200} // default 200 response code
		defer instance.buffer.Reset()

		next.ServeHTTP(instance, r)

		size, e := instance.Done()
		if e != nil {
			slog.ErrorContext(ctx, "Error Writing Response", slog.String("error", e.Error()))
		}

		slog.InfoContext(ctx, "Response", slog.String("path", r.URL.Path), slog.Int64("size", size), slog.Int("status", instance.status), slog.String("content-type", w.Header().Get("Content-Type")))
	})
}

func (w *Writer) Header() http.Header {
	return w.w.Header()
}

func (w *Writer) Write(bytes []byte) (int, error) {
	return w.buffer.Write(bytes)
}

func (w *Writer) WriteHeader(status int) {
	w.status = status
}

func (w *Writer) Done() (int64, error) {
	if w.status >= 100 {
		w.w.WriteHeader(w.status)
	}

	return io.Copy(w.w, &w.buffer)
}
