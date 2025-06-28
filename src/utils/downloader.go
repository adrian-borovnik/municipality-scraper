package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

func normalizeFileName(fileName string) string {
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, "č", "c")
	fileName = strings.ReplaceAll(fileName, "š", "s")
	fileName = strings.ReplaceAll(fileName, "ž", "z")
	fileName = strings.ReplaceAll(fileName, "–", "")
	fileName = strings.ReplaceAll(fileName, "-", "")
	fileName = strings.ReplaceAll(fileName, "  ", "_")
	fileName = strings.ReplaceAll(fileName, " ", "_")
	return fileName
}

func downloadImage(url string, outFolder string, fileName string) error {
	maxRetries := 5
	baseDelay := time.Second

	var res *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err = http.Get(url)
		if err != nil {
			fmt.Println("Failed to download image:", err)
			time.Sleep(baseDelay * (1 << attempt))
			continue
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			break
		} else if res.StatusCode == 429 {
			fmt.Println("Rate limit exceeded")
			waitTime := baseDelay * (1 << attempt)
			jitter := time.Duration(rand.Intn(1000)) * time.Microsecond
			time.Sleep(waitTime + jitter)
			continue
		} else {
			return fmt.Errorf("failed: %s", res.Status)
		}
	}

	if res == nil || res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s after retries", url)
	}

	splitUrl := strings.Split(url, ".")
	fileExtension := splitUrl[len(splitUrl)-1]
	fileName = normalizeFileName(fileName)
	path := fmt.Sprintf("%s/%s.%s", outFolder, fileName, fileExtension)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadMunicipalityImages(municipalities []*Municipality, outFolder string, goRoutinesCount int64) {
	err := os.MkdirAll(outFolder, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	sem := semaphore.NewWeighted(goRoutinesCount)

	for _, m := range municipalities {
		if m.ImgUrl == "" {
			continue
		}

		wg.Add(1)
		go func(m *Municipality) {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				fmt.Printf("Failed to acquire semaphore for %s: %v\n", m.Name, err)
				return
			}
			defer sem.Release(1)

			time.Sleep(time.Second * time.Duration(rand.Intn(5)))
			err := downloadImage(m.ImgUrl, outFolder, m.Name)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", m.Name, err)
			}

			fmt.Println("Successfully downloaded image for", m.Name)
		}(m)
	}
	wg.Wait()
}

func LoadMunicipalitiesFromFile(filePath string) ([]*Municipality, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var municipalities []*Municipality
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&municipalities); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return municipalities, nil
}
