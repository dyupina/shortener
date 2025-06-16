package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	models "shortener/internal/domain/models/json"
	"syscall"
	"time"
)

var shorturl struct {
	URL string `json:"result"`
}

var origurl struct {
	URL string `json:"url"`
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write overrides the Write method of the http.ResponseWriter interface.
// The function writes data to the HTTP response and updates the size of
// written data in the responseData structure for subsequent logging.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader overrides the WriteHeader method of the http.ResponseWriter interface.
//
// The function writes the status code to the HTTP response and updates it
// in the responseData structure for subsequent logging.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// gzipWriter wraps http.ResponseWriter to support data compression using gzip.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write writes compressed data to the HTTP response.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func extractURLfromHTML(res http.ResponseWriter, req *http.Request) string {
	b, _ := io.ReadAll(req.Body)
	body := string(b)

	re := regexp.MustCompile(`href=['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(body)

	if len(matches) > 1 {
		return matches[1]
	} else {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return ""
	}
}

func extractURLfromJSON(res http.ResponseWriter, req *http.Request) string {
	if err := json.NewDecoder(req.Body).Decode(&origurl); err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return ""
	}
	return origurl.URL
}

func extractURLsfromJSONBatchRequest(req *http.Request) []models.BatchRequestEntity {
	var urls []models.BatchRequestEntity
	err := json.NewDecoder(req.Body).Decode(&urls)
	if err != nil {
		return nil
	}
	return urls
}

// HandleGracefulShutdown handles termination signals.
func (con *Controller) HandleGracefulShutdown(server *http.Server) {
	notifyCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Ждем получения первого сигнала
	<-notifyCtx.Done()
	con.Logger.Infof("Received shutdown signal")

	// Отключаем прием новых подключений и дожидаемся завершения активных запросов
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(con.Config.Timeout)*time.Second)
	defer cancel()

	// Закрываем соединение с базой данных.
	go func() {
		if con.Config.DBConnection != "" {
			con.Logger.Infof("Closing database connection...")
			if err := con.StorageService.Close(); err != nil {
				con.Logger.Errorf("Failed to close database connection: %v", err)
			}
		}
	}()

	con.Logger.Infof("Shutting down gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		con.Logger.Infof("HTTP server shutdown error: %v", err)
	}

	con.Logger.Infof("Server has been shut down.")
}

// IsIPInSubnet checks if an IP address is in the specified subnet (CIDR)
func (con *Controller) IsIPInSubnet(ip, subnet string) bool {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		con.Logger.Infof("Invalid subnet format: %v", err)
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		con.Logger.Infof("Invalid IP address format: %s", ip)
		return false
	}

	return ipNet.Contains(parsedIP)
}
