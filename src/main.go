package main

import (
	"fmt"
	"municipality-scrapper/src/utils"
)

const (
	protocol    = "https"
	domain      = "sl.wikipedia.org"
	baseUrl     = protocol + "://" + domain
	homePageUrl = baseUrl + "/wiki/Seznam_ob%C4%8Din_v_Sloveniji"
	outFolder   = "./data"

	imageDownloadGoRoutinesCount int64 = 5
)

func main() {
	municipalities := utils.ScrapeMunicipalityData(protocol, domain, baseUrl, homePageUrl)

	fmt.Println("Municipalities found:", len(municipalities))

	utils.DownloadMunicipalityImages(municipalities, outFolder, imageDownloadGoRoutinesCount)
}
