package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type Municipality struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	WikiUrl    string `json:"url"`
	ImgPageUrl string `json:"imgRedirectUrl"`
	ImgUrl     string `json:"imgUrl"`
}

func ScrapeMunicipalityData(protocol, baseUrl, homePageUrl string) map[int]*Municipality {
	var mutex sync.RWMutex
	var municipalities = make(map[int]*Municipality)
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

		info := &Municipality{
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

		imgRedirectUrl := baseUrl + href

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		m := municipalities[idInt]
		m.ImgPageUrl = imgRedirectUrl
		mutex.Unlock()

	})

	imageCollector.OnHTML("#bodyContent .fullMedia .internal", func(e *colly.HTMLElement) {
		ctx := e.Request.Ctx
		id := ctx.Get("id")

		imgUrl := protocol + ":" + e.Attr("href")

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		m := municipalities[idInt]
		m.ImgUrl = imgUrl
		mutex.Unlock()
	})

	fmt.Println("Searching the wiki websites")
	err := c.Visit(homePageUrl)
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

	return municipalities
}
