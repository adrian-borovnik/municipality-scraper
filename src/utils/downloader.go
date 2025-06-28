package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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

func downloadImage(wg *sync.WaitGroup, url string, outFolder string, fileName string) error {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadMunicipalityImages(municipalities map[int]*Municipality, outFolder string) {
	fmt.Println("Downloading the coa images")

	err := os.MkdirAll(outFolder, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	for _, m := range municipalities {
		if m.ImgUrl == "" {
			continue
		}

		wg.Add(1)
		go func(m *Municipality) {
			time.Sleep(time.Second * time.Duration(rand.Intn(5)))
			err := downloadImage(&wg, m.ImgUrl, outFolder, m.Name)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", m.Name, err)
			}
		}(m)
	}
	wg.Wait()

}
