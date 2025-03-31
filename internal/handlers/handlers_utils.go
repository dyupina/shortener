package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"sync"
)

var shorturl struct {
	URL string `json:"result"`
}

var origurl struct {
	URL string `json:"url"`
}

type batchRequestEntity struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponseEntity struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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

func extractURLsfromJSONBatchRequest(req *http.Request) []batchRequestEntity {
	var urls []batchRequestEntity
	err := json.NewDecoder(req.Body).Decode(&urls)
	if err != nil {
		return nil
	}
	return urls
}

func createURLBatchChannel(doneCh chan struct{}, urlsToDeleteArray []string) chan []string {
	inputCh := make(chan []string)
	go func() {
		defer close(inputCh)
		select {
		case <-doneCh:
			return
		case inputCh <- urlsToDeleteArray:
		}
	}()
	return inputCh
}

func distributeDeleteTasks(doneCh chan struct{}, inputCh chan []string, numWorkers int, userID string, con *Controller) []chan string {
	var resultChs []chan string
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		resultCh := make(chan string)

		wg.Add(1)
		go func(ch chan string) {
			defer wg.Done()
			defer close(ch)
			for urlsToDeleteArray := range inputCh {
				select {
				case <-doneCh:
					return
				default:
					err := con.storageService.BatchDeleteURLs(userID, urlsToDeleteArray)
					if err != nil {
						con.sugar.Errorf(" Error Updating flag to URLs %s\n", err.Error())

						close(doneCh)
						return
					}

					for _, d := range urlsToDeleteArray {
						ch <- d
					}
				}
			}
		}(resultCh)

		resultChs = append(resultChs, resultCh)
	}

	go func() {
		wg.Wait()
	}()

	return resultChs
}

func collectDeletionResults(channels ...chan string) chan string {
	finalCh := make(chan string)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(ch chan string) {
			defer wg.Done()
			for v := range ch {
				finalCh <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
