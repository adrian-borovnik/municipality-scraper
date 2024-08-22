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
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(resp.Status)
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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", fileName, fileExtension), nil
}

type MunicipalityInfo struct {
	Id   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Img  string `json:"img"`
}

func main() {
	const UrlStr = "https://en.wikipedia.org/wiki/Municipalities_of_Slovenia"
	const CtxKey = "info"
	var municipalities []MunicipalityInfo

	c := colly.NewCollector(
	//colly.Async(true),
	)

	coaCollector := c.Clone()
	imageCollector := c.Clone()

	//c.OnHTML(".wikitable tbody tr > td:first-child", func(e *colly.HTMLElement) {
	//	URL, err := url2.Parse(UrlStr)
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//
	//	if e.Request.URL.Host != URL.Host || e.Request.URL.Path != URL.Path {
	//		return
	//	}
	//
	//
	//})

	c.OnHTML(".wikitable td a", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "Municipality") {
			return
		}

		name := strings.ReplaceAll(e.Text, "Municipality of ", "")
		name = strings.ReplaceAll(name, "Urban ", "")

		info := &MunicipalityInfo{
			Id:   len(municipalities),
			Name: name,
		}

		url := e.Request.AbsoluteURL(e.Attr("href"))
		e.Response.Ctx.Put(CtxKey, info)
		err := coaCollector.Request("GET", url, nil, e.Response.Ctx, nil)
		if err != nil {
			fmt.Println(err)
		}

		municipalities = append(municipalities, *info)

	})

	coaCollector.OnHTML(".infobox a[title]", func(e *colly.HTMLElement) {
		title := e.Attr("title")
		if !strings.Contains(title, "Coat") {
			return
		}

		url := e.Request.AbsoluteURL(e.Attr("href"))

		// NOTE: e.Response.Ctx is being passed forward from the main collector to the imageCollector
		err := imageCollector.Request("GET", url, nil, e.Response.Ctx, nil)
		if err != nil {
			fmt.Println(err)
		}

	})

	imageCollector.OnHTML(".fullMedia > p > a", func(e *colly.HTMLElement) {
		if info, ok := e.Response.Ctx.GetAny(CtxKey).(*MunicipalityInfo); ok == true {

			fmt.Printf("Downloading COA of Municipality of %s\n", info.Name)

			url := e.Attr("href")
			imgName, err := downloadImage(e.Request.AbsoluteURL(url), "./out", info.Name)
			if err != nil {
				fmt.Println(err)
			}

			info.Img = imgName
		}
	})

	err := c.Visit(UrlStr)
	if err != nil {
		fmt.Println(err)
	}

	//c.Wait()

	fmt.Println(len(municipalities))

	jsonBytes, err := json.Marshal(municipalities)
	if err != nil {
		fmt.Println(err)
	}

	err = os.WriteFile("./out/_data.json", jsonBytes, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
