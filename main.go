package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"os"
	"strings"
)

func downloadImage(url string, outFolder string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(res.Status)
	}

	splitUrl := strings.Split(url, "/")
	fileName := splitUrl[len(splitUrl)-1]

	file, err := os.Create(outFolder + "/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	return err
}

func main() {

	var municipalities []string

	c := colly.NewCollector()

	c.OnHTML(".wikitable td a", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "Municipality") {
			return
		}

		err := c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		if err != nil {
			fmt.Println(err)
		}

		municipality := strings.ReplaceAll(e.Text, "Municipality of ", "")
		municipality = strings.ReplaceAll(municipality, "Urban ", "")
		municipalities = append(municipalities, municipality)
		fmt.Println(municipality)
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

	c.OnHTML(".fullImageLink img[src]", func(e *colly.HTMLElement) {
		alt := e.Attr("alt")
		if !strings.Contains(alt, "File:") {
			return
		}

		url := e.Attr("src")
		err := downloadImage(e.Request.AbsoluteURL(url), "./out")
		if err != nil {
			fmt.Println(err)
		}
	})

	//c.OnRequest(func(r *colly.Request) {
	//	fmt.Println(r.URL)
	//})

	err := c.Visit("https://en.wikipedia.org/wiki/Municipalities_of_Slovenia")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(municipalities))
}
