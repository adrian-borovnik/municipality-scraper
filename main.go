package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"os"
	"strings"
)

func normalizeFileName(fileName string) string {
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, "č", "c")
	fileName = strings.ReplaceAll(fileName, "š", "s")
	fileName = strings.ReplaceAll(fileName, "ž", "z")
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "–", "_")
	fileName = strings.ReplaceAll(fileName, "-", "_")
	return fileName
}

func downloadImage(url string, outFolder string, fileName string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(res.Status)
	}

	splitUrl := strings.Split(url, ".")
	fileExtension := splitUrl[len(splitUrl)-1]

	fileName = normalizeFileName(fileName)

	path := fmt.Sprintf("%s/%s.%s", outFolder, fileName, fileExtension)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}

type Municipality struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func main() {

	var municipalities []Municipality
	currentMunicipality := ""

	c := colly.NewCollector()

	c.OnHTML(".wikitable td a", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "Municipality") {
			return
		}

		municipality := strings.ReplaceAll(e.Text, "Municipality of ", "")
		municipality = strings.ReplaceAll(municipality, "Urban ", "")
		currentMunicipality = municipality

		err := c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		if err != nil {
			fmt.Println(err)
		}
	})

	c.OnHTML(".infobox a[title]", func(e *colly.HTMLElement) {
		title := e.Attr("title")
		if !strings.Contains(title, "Coat") {
			return
		}

		url := e.Request.AbsoluteURL(e.Attr("href"))
		err := c.Visit(url)
		if err != nil {
			fmt.Println(err)
		}
	})

	c.OnHTML(".fullMedia > p > a", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		fmt.Println("Downloading COA of Municipality of " + currentMunicipality)
		imgPath, err := downloadImage(e.Request.AbsoluteURL(url), "./out", currentMunicipality)
		if err != nil {
			fmt.Println(err)
		}

		municipalities = append(municipalities, Municipality{
			Name: currentMunicipality,
			Path: imgPath,
		})
	})

	//c.OnRequest(func(r *colly.Request) {
	//	fmt.Println(r.URL)
	//})

	err := c.Visit("https://en.wikipedia.org/wiki/Municipalities_of_Slovenia")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(municipalities))
	jsonBytes, err := json.Marshal(municipalities)
	if err != nil {
		fmt.Println(err)
	}

	if err := os.WriteFile("./out/_data.json", jsonBytes, 0644); err != nil {
		fmt.Println(err)
	}
}
