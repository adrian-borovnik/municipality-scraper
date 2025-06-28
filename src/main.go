package main

const (
	protocol    = "https"
	baseUrl     = protocol + "://sl.wikipedia.org"
	homePageUrl = baseUrl + "/wiki/Seznam_ob%C4%8Din_v_Sloveniji"
	outFolder   = "./data"
)

func main() {
	municipalities := ScrapeMunicipalityData(protocol, baseUrl, homePageUrl)
	DownloadMunicipalityImages(municipalities, outFolder)
}
