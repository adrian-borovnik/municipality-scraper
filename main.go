package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
)

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

	c.OnHTML(".infobox img", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Attr("alt"), "Coat") {
			return
		}

		fmt.Println(e.Attr("alt"))
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
