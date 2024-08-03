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
		municipality := strings.ReplaceAll(e.Text, "Municipality of ", "")
		municipality = strings.ReplaceAll(municipality, "Urban ", "")
		municipalities = append(municipalities, municipality)
		fmt.Println(municipality)
	})

	err := c.Visit("https://en.wikipedia.org/wiki/Municipalities_of_Slovenia")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(municipalities))
}
