package handlers

import (
	"net/http"
	"time"

	"example.com/tasksapi/models"
)

// LoggingMiddleware intercepta todas as requisições HTTP e loga informações
type LoggingMiddleware struct {
	logger models.Logger
}

// NewLoggingMiddleware cria um middleware de logging
func NewLoggingMiddleware(logger models.Logger) *LoggingMiddleware {
	if logger == nil {
		logger = models.NewDefaultLogger()
	}
	return &LoggingMiddleware{logger: logger}
}

// responseWriter wrapper para capturar o status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// Middleware retorna um handler que loga todas as requisições
func (lm *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log da requisição recebida com cor baseada no método
		methodColor := getMethodColor(r.Method)
		lm.logger.Info("[HTTP] %s%s%s %s - Started", methodColor, r.Method, models.ColorReset, r.RequestURI)

		// Wrapper para capturar status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Processa a requisição
		next.ServeHTTP(wrapped, r)

		// Log da resposta baseado no status code
		duration := time.Since(start)
		logResponse(lm.logger, r.Method, r.RequestURI, wrapped.statusCode, duration, wrapped.written)
	})
}

// getMethodColor retorna a cor ANSI baseada no método HTTP
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return models.ColorBlue
	case "POST":
		return models.ColorGreen
	case "PUT":
		return models.ColorYellow
	case "DELETE":
		return models.ColorRed
	default:
		return models.ColorWhite
	}
}

// logResponse loga a resposta com cor baseada no status code
func logResponse(logger models.Logger, method, uri string, statusCode int, duration time.Duration, bytes int) {
	msg := "[HTTP] %s %s - Completed %s%d%s in %v (%d bytes)"
	statusColor := getStatusColor(statusCode)

	switch {
	case statusCode >= 500:
		logger.Error(msg, method, uri, statusColor, statusCode, models.ColorReset, duration, bytes)
	case statusCode >= 400:
		logger.Warn(msg, method, uri, statusColor, statusCode, models.ColorReset, duration, bytes)
	default:
		logger.Info(msg, method, uri, statusColor, statusCode, models.ColorReset, duration, bytes)
	}
}

// getStatusColor retorna a cor ANSI baseada no status code HTTP
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 500:
		return models.ColorRed + models.ColorBold
	case statusCode >= 400:
		return models.ColorYellow
	case statusCode >= 300:
		return models.ColorCyan
	case statusCode >= 200:
		return models.ColorGreen
	default:
		return models.ColorWhite
	}
}
