// Package services contains the implementation of URLserv.
package services

import (
	"errors"
	"fmt"
	"shortener/internal/config"
	models "shortener/internal/domain/models/json"
	"shortener/internal/repository"
	"shortener/internal/storage"
	"sync"
)

type URLserv struct {
	config         *config.Config
	storageService storage.StorageService
	userService    UserService
}

type URLService interface {
	APIGetUserURLs(userID string) ([]UserURL, bool)
	ShortenURL(originalURL, userID string) (string, error)
	APIShortenBatchURL(userID string, urls []models.BatchRequestEntity) ([]models.BatchResponseEntity, error)
	GettingOriginalURL(shortID string) (originalURL string, isDeleted bool, err error)
	PingHandler() error
	Statistics() models.StatsResponse
	DeleteUserURLs(userID string, urlIDs []string) (<-chan string, error)
	DistributeDeleteTasks(doneCh chan struct{}, inputCh chan []string, numWorkers int, userID string) []chan string
}

func NewURLService(conf *config.Config, storageService storage.StorageService, userService UserService) URLService {
	return &URLserv{
		config:         conf,
		storageService: storageService,
		userService:    userService,
	}
}

func (s *URLserv) ShortenURL(originalURL, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("Unauthorized")
	}

	shortID, err := s.storageService.UpdateData(originalURL, userID)
	if err != nil && errors.Is(err, repository.ErrDuplicateURL) {
		return shortID, repository.ErrDuplicateURL
	}

	s.userService.AddURLs(s.config.BaseURL, userID, shortID, originalURL)

	return shortID, nil
}

func (s *URLserv) APIShortenBatchURL(userID string, urls []models.BatchRequestEntity) ([]models.BatchResponseEntity, error) {
	if userID == "" {
		return nil, fmt.Errorf("Unauthorized")
	}

	batchResponse := []models.BatchResponseEntity{}
	var errUpdateData error

	for _, url := range urls {
		shortID, err := s.storageService.UpdateData(url.OriginalURL, userID)
		errUpdateData = err

		if err == nil {
			s.userService.AddURLs(s.config.BaseURL, userID, shortID, url.OriginalURL)
		}

		batchResponse = append(batchResponse, models.BatchResponseEntity{
			CorrelationID: url.CorrelationID,
			ShortURL:      s.config.BaseURL + "/" + shortID,
		})
	}

	if errUpdateData != nil && errors.Is(errUpdateData, repository.ErrDuplicateURL) {
		return batchResponse, repository.ErrDuplicateURL
	}

	return batchResponse, nil
}

func (s *URLserv) GettingOriginalURL(shortID string) (originalURL string, isDeleted bool, err error) {
	originalURL, isDeleted, err = s.storageService.GetData(shortID)
	if err != nil {
		return "", false, errors.New("failed to retrieve original URL")
	}
	if isDeleted {
		return "", true, nil
	}
	return originalURL, false, nil
}

func (s *URLserv) PingHandler() error {
	err := s.storageService.Ping()
	if err != nil {
		return errors.New("database connection error")
	}
	return nil
}

func (s *URLserv) Statistics() models.StatsResponse {
	return models.StatsResponse{
		URLs:  s.userService.GetURLsCount(),
		Users: s.userService.GetUserNumber(),
	}
}

func (s *URLserv) APIGetUserURLs(userID string) ([]UserURL, bool) {
	urls, exist := s.userService.GetUserURLs(userID)
	if !exist {
		return nil, false
	}
	return urls, true
}

func (s *URLserv) DeleteUserURLs(userID string, urlIDs []string) (<-chan string, error) {
	if userID == "" {
		return nil, fmt.Errorf("Unauthorized")
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	inputCh := createURLBatchChannel(doneCh, urlIDs)
	workerChs := s.DistributeDeleteTasks(doneCh, inputCh, s.config.NumWorkers, userID)
	resultCh := collectDeletionResults(workerChs...)

	return resultCh, nil
}

func (s *URLserv) DistributeDeleteTasks(doneCh chan struct{}, inputCh chan []string, numWorkers int, userID string) []chan string {
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
					err := s.storageService.BatchDeleteURLs(userID, urlsToDeleteArray)
					if err != nil {
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
