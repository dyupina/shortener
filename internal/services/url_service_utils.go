package services

import "sync"

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
