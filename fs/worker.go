package fs

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// ProcessFilesParallel applies process to each file path using the given
// number of concurrent workers. When workers is 1 it runs sequentially.
// Workers are clamped to min(workers, len(files), NumCPU) to avoid
// spawning more goroutines than necessary.
func ProcessFilesParallel(files []string, workers int, process func(path string) error) error {
	if workers <= 0 {
		workers = 1
	}
	if workers > len(files) {
		workers = len(files)
	}
	if max := runtime.NumCPU(); workers > max {
		workers = max
	}

	if workers <= 1 {
		for _, f := range files {
			if err := process(f); err != nil {
				return err
			}
		}
		return nil
	}

	ch := make(chan string, workers)
	var mu sync.Mutex
	var errs []string
	var wg sync.WaitGroup

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for path := range ch {
				if err := process(path); err != nil {
					mu.Lock()
					errs = append(errs, err.Error())
					mu.Unlock()
				}
			}
		}()
	}

	for _, f := range files {
		ch <- f
	}
	close(ch)

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("errors processing files:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}
