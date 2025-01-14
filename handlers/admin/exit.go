package admin

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kevin07696/receipt-processor/domain"
)

func Exit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		path = strings.Trim(path, "/")
		segments := strings.Split(path, "/")
		param := segments[1]

		code, err := strconv.Atoi(param)
		if err != nil {
			slog.Debug("StatusBadRequest: uuid is invalid", slog.String("code", param), slog.Any("error", err))
			http.Error(w, domain.ErrorToCodes[domain.ErrBadRequest].Message, domain.ErrorToCodes[domain.ErrBadRequest].Code)
			return
		}

		slog.Debug(fmt.Sprintf("exit code %d", code))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

		// Flush the response to ensure it is sent to the client
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		go func(exitCode int) {
			os.Exit(exitCode)
		}(code)
	}
}
