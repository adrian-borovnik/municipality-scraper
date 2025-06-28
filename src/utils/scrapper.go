package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func ScrapeMunicipalityData(protocol, domain, baseUrl, homePageUrl string) []*Municipality {
	var mutex sync.RWMutex
	municipalitiesMap := make(map[int]*Municipality)
	count := 0

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(domain),
	)

	coaCollector := c.Clone()
	imageCollector := c.Clone()

	c.OnHTML(".wikitable tbody tr>td:nth-child(2) a", func(e *colly.HTMLElement) {
		name := e.Text
		wikiUrl := e.Request.AbsoluteURL(e.Attr("href"))

		mutex.Lock()
		defer mutex.Unlock()

		count++
		info := &Municipality{
			Id:      count,
			Name:    name,
			WikiUrl: wikiUrl,
		}
		municipalitiesMap[count] = info
	})

	coaCollector.OnHTML(".infobox .infobox-full-data>table>tbody span>a[title]", func(e *colly.HTMLElement) {
		ctx := e.Request.Ctx
		id := ctx.Get("id")

		title := e.Attr("title")
		if strings.Contains(title, "Zastava") {
			return
		}

		imgRedirectUrl := baseUrl + e.Attr("href")

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		if m, ok := municipalitiesMap[idInt]; ok {
			m.ImgPageUrl = imgRedirectUrl
		}
		mutex.Unlock()
	})

	imageCollector.OnHTML("#bodyContent .fullMedia .internal", func(e *colly.HTMLElement) {
		ctx := e.Request.Ctx
		id := ctx.Get("id")

		imgUrl := protocol + ":" + e.Attr("href")

		idInt, _ := strconv.Atoi(id)
		mutex.Lock()
		if m, ok := municipalitiesMap[idInt]; ok {
			m.ImgUrl = imgUrl
		}
		mutex.Unlock()
	})

	fmt.Println("Searching the wiki websites")
	if err := c.Visit(homePageUrl); err != nil {
		log.Fatal(err)
	}
	c.Wait()

	fmt.Println("Visiting the municipalities pages to get the coa imgPageUrl")
	for _, m := range municipalitiesMap {
		ctx := colly.NewContext()
		ctx.Put("id", strconv.Itoa(m.Id))
		_ = coaCollector.Request("GET", m.WikiUrl, nil, ctx, nil)
	}
	coaCollector.Wait()

	fmt.Println("Visiting the municipalities pages to get the coa imgUrl")
	for _, m := range municipalitiesMap {
		if m.ImgPageUrl == "" {
			continue
		}
		ctx := colly.NewContext()
		ctx.Put("id", strconv.Itoa(m.Id))
		_ = imageCollector.Request("GET", m.ImgPageUrl, nil, ctx, nil)
	}
	imageCollector.Wait()

	// Convert to slice
	municipalitySlice := make([]*Municipality, 0, len(municipalitiesMap))
	for _, m := range municipalitiesMap {
		municipalitySlice = append(municipalitySlice, m)
	}

	return municipalitySlice
}

func SaveMunicipalityDataAsJson(municipalities []*Municipality, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(municipalities); err != nil {
		fmt.Printf("Failed to encode JSON: %v\n", err)
		return
	}

	fmt.Printf("Saved data to %s\n", filePath)
}
