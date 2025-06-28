package main

import (
	"fmt"
	"municipality-scrapper/src/utils"
	"os"
)

const (
	protocol    = "https"
	domain      = "sl.wikipedia.org"
	baseUrl     = protocol + "://" + domain
	homePageUrl = baseUrl + "/wiki/Seznam_ob%C4%8Din_v_Sloveniji"

	jsonDataFilePath = "./data.json"

	imagesOutFolder                    = "./coat_of_arms"
	imageDownloadGoRoutinesCount int64 = 10
)

const (
	scrapeCommand   = "scrape"
	downloadCommand = "download"
	helpCommand     = "help"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("No command provided.")
		printHelp()
		return
	}

	command := args[0]
	arguments := args[1:]

	switch command {
	case scrapeCommand:
		filePath := jsonDataFilePath // default path

		if len(arguments) >= 1 && arguments[0] != "" {
			filePath = arguments[0]
		}

		fmt.Printf("Scraping municipality data from %s...\n", homePageUrl)
		municipalities := utils.ScrapeMunicipalityData(protocol, domain, baseUrl, homePageUrl)
		fmt.Println("Municipalities found:", len(municipalities))

		utils.SaveMunicipalityDataAsJson(municipalities, filePath)

	case downloadCommand:
		if len(arguments) < 1 {
			fmt.Println("Usage: municipality-scrapper download <input_file> [output_folder]")
			fmt.Println("  <input_file> is required")
			fmt.Println("  [output_folder] is optional (default: ./coat_of_arms)")
			os.Exit(1)
		}

		originFilePath := arguments[0]
		destinationFolder := imagesOutFolder // default

		if len(arguments) >= 2 && arguments[1] != "" {
			destinationFolder = arguments[1]
		}

		fmt.Println("Loading municipalities from", originFilePath)
		municipalities, err := utils.LoadMunicipalitiesFromFile(originFilePath)
		if err != nil {
			fmt.Println("Error loading municipalities:", err)
			os.Exit(1)
		}

		fmt.Println("Downloading municipality coat of arms images...")
		utils.DownloadMunicipalityImages(municipalities, destinationFolder, imageDownloadGoRoutinesCount)

	case helpCommand:
		printHelp()

	default:
		fmt.Println("Invalid command:", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Municipality Scraper - Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  municipality-scrapper <command> [arguments]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("  scrape [output_file]")
	fmt.Println("    Scrapes municipality data from Wikipedia and saves it as JSON.")
	fmt.Println("    - output_file: (optional) Path to save the JSON data. Default: ./data.json")
	fmt.Println()
	fmt.Println("  download <input_file> [output_folder]")
	fmt.Println("    Downloads coat of arms images using municipality JSON data.")
	fmt.Println("    - input_file:       Path to the JSON file containing municipality data.")
	fmt.Println("    - output_folder:    (optional) Destination folder for images. Default: ./coat_of_arms")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Shows this help message.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  municipality-scrapper scrape municipalities.json")
	fmt.Println("  municipality-scrapper download municipalities.json ./images")
	fmt.Println("  municipality-scrapper help")
}
