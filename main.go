package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

func main() {
	fmt.Println("Adrian was here!")

	c := colly.NewCollector(
		colly.AllowedDomains("https://sl.wikipedia.org/wiki/Seznam_ob%C4%8Din_v_Sloveniji"),
	)

	fmt.Println(c)
}
