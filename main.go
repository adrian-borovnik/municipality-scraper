package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
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

type MunicipalityInfo struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	WikiUrl    string `json:"url"`
	ImgPageUrl string `json:"imgRedirectUrl"`
	ImgUrl     string `json:"imgUrl"`
}

const Protocol = "https"
const BaseUrl = Protocol + "://sl.wikipedia.org"
const HomePageUrl = BaseUrl + "/wiki/Seznam_ob%C4%8Din_v_Sloveniji"

func main() {
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
		wikiUrl := e.Request.AbsoluteURL(e.Attr("href"))

		mutex.Lock()
		count++

		info := &MunicipalityInfo{
			Id:      count,
			Name:    name,
			WikiUrl: wikiUrl,
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
		m.ImgPageUrl = imgRedirectUrl
		mutex.Unlock()

		// url := e.Request.AbsoluteURL(e.Attr("href"))

		// // NOTE: e.Response.Ctx is being passed forward from the main collector to the imageCollector
		// err := imageCollector.Request("GET", url, nil, e.Response.Ctx, nil)
		// if err != nil {
		// 	fmt.Println(err)
		// }

	})

	imageCollector.OnHTML("#bodyContent .fullMedia .internal", func(e *colly.HTMLElement) {
		ctx := e.Request.Ctx
		id := ctx.Get("id")

		imgUrl := Protocol + ":" + e.Attr("href")

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		m := municipalities[idInt]
		m.ImgUrl = imgUrl
		mutex.Unlock()

		// if info, ok := e.Response.Ctx.GetAny(CtxKey).(*MunicipalityInfo); ok == true {

		// 	fmt.Printf("Downloading COA of Municipality of %s\n", info.Name)

		// 	url := e.Attr("href")
		// 	imgName, err := downloadImage(e.Request.AbsoluteURL(url), "./out", info.Name)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 	}

		// 	info.Img = imgName
		// }
	})

	fmt.Println("Searching the wiki websites")
	err := c.Visit(HomePageUrl)
	if err != nil {
		log.Fatal(err)
	}
	c.Wait()

	fmt.Println("Visiting the municipalities pages to get the coa imgPageUrl")
	for _, m := range municipalities {
		ctx := colly.NewContext()
		ctx.Put("id", strconv.Itoa(m.Id))

		err := coaCollector.Request("GET", m.WikiUrl, nil, ctx, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	coaCollector.Wait()

	fmt.Println("Visiting the municipalities pages to get the coa imgUrl")
	for _, m := range municipalities {
		if m.ImgPageUrl == "" {
			continue
		}

		ctx := colly.NewContext()
		ctx.Put("id", strconv.Itoa(m.Id))

		err := imageCollector.Request("GET", m.ImgPageUrl, nil, ctx, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	imageCollector.Wait()

	// TODO Save the municipalities to a file

	// TODO Add arguments for getting the data or to download images
	// Add wait latency for 429 Error

	// return

	fmt.Println("Downloading the coa images")

	var wg sync.WaitGroup
	for _, m := range municipalities {
		if m.ImgUrl == "" {
			continue
		}

		wg.Add(1)
		go func(m *MunicipalityInfo) {
			time.Sleep(time.Second * time.Duration(rand.Intn(5)))
			err := downloadImage(&wg, m.ImgUrl, "./out", m.Name)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", m.Name, err)
			}
		}(m)
	}
	wg.Wait()

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
