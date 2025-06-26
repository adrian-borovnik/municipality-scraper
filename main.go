package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
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
	Id             int    `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Url            string `json:"url"`
	ImgRedirectUrl string `json:"imgRedirectUrl"`
	Img            string `json:"img"`
}

func main() {
	fmt.Println("Searching for coat of arms")

	const BaseUrl = "https://sl.wikipedia.org"
	const HomePageUrl = BaseUrl + "/wiki/Seznam_ob%C4%8Din_v_Sloveniji"
	const CtxKey = "info"

	var mutex sync.RWMutex
	var municipalities = make(map[int]*MunicipalityInfo)
	count := 0

	c := colly.NewCollector(
		colly.Async(true),
	)

	coaCollector := c.Clone()
	imageCollector := c.Clone()

	c.OnHTML(".wikitable tbody tr>td:nth-child(2) a", func(e *colly.HTMLElement) {
		name := e.Text
		url := e.Request.AbsoluteURL(e.Attr("href"))

		mutex.Lock()
		count++

		info := &MunicipalityInfo{
			Id:   count,
			Name: name,
			Url:  url,
		}

		municipalities[count] = info
		mutex.Unlock()
	})

	coaCollector.OnHTML(".infobox .infobox-full-data>table>tbody span>a[title]", func(e *colly.HTMLElement) {
		ctx := e.Request.Ctx
		id := ctx.Get("id")

		title := e.Attr("title")
		if strings.Contains(title, "Zastava") {
			return
		}

		href := e.Attr("href")

		imgRedirectUrl := BaseUrl + href

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		m := municipalities[idInt]
		m.ImgRedirectUrl = imgRedirectUrl
		mutex.Unlock()

		// url := e.Request.AbsoluteURL(e.Attr("href"))

		// // NOTE: e.Response.Ctx is being passed forward from the main collector to the imageCollector
		// err := imageCollector.Request("GET", url, nil, e.Response.Ctx, nil)
		// if err != nil {
		// 	fmt.Println(err)
		// }

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

	err := c.Visit(HomePageUrl)
	if err != nil {
		fmt.Println(err)
	}
	c.Wait()

	for _, m := range municipalities {
		ctx := colly.NewContext()
		ctx.Put("id", strconv.Itoa(m.Id))

		err := coaCollector.Request("GET", m.Url, nil, ctx, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	coaCollector.Wait()

	for _, m := range municipalities {
		fmt.Println(m.Id, m.Name, m.ImgRedirectUrl)
	}

	// fmt.Println(len(municipalities))

	// jsonBytes, err := json.Marshal(municipalities)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// err = os.WriteFile("./out/_data.json", jsonBytes, 0644)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
