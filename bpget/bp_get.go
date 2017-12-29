package main

import (
	"github.com/PuerkitoBio/goquery"
	"encoding/csv"
	"os"
	"log"
	"strings"
	"fmt"
	"time"
	"flag"
)

type BluePrint struct {
	Priority string
	Title    string
	Series   string
	Url      string
}

func GetBlueprint(service string, version string) []*BluePrint {
	baseURL := "https://blueprints.launchpad.net"
	docURL := strings.Join([]string{baseURL, service, version}, "/")
	fmt.Println("Donwload blueprints from: " + docURL)
	doc, err := goquery.NewDocument(docURL)
	if err != nil {
		panic(err)
	}
	var l []*BluePrint
	// Find the bp table in html
	doc.Find("#speclisting tbody tr").Each(func(i int, bpRow *goquery.Selection) {
		// Get col in every row
		bpCol := bpRow.Find("td")
		// Get priority
		bp := new(BluePrint)
		bp.Priority = bpCol.Eq(0).Find("span").Eq(1).Text()
		// Get url and title
		Name := bpCol.Eq(1).Find("a")
		url, exist := Name.Attr("href")
		if exist {
			bp.Url = baseURL + url
		}
		title, exist := Name.Attr("title")
		if exist {
			bp.Title = title
		}
		// Get version
		bp.Series = bpCol.Eq(5).Find("a").Text()
		if bp.Series == "" {
			bp.Series = version
		}
		// append bp on l
		l = append(l, bp)
	})

	return l
}

func getFileName(s string, v string, l string) string {
	t := time.Now().Format("2006_01_02_15_04_05")
	var fileName string
	if l != "" {
		fileName = l + strings.Join([]string{s, v, t}, "_")
	} else {
		fileName = strings.Join([]string{s, v, t}, "_")
	}

	return fileName + ".csv"
}

func main() {
	servicePtr := flag.String("service", "nova", "The service name of OpenStack (nova, cinder...)")
	versionPtr := flag.String("version", "queens", "The version of service (pike, queens...)")
	locationPtr := flag.String("location", "", "The location of the output file. (D:\\folder, ..\\folder)")
	flag.Parse()
	service := *servicePtr
	version := *versionPtr
	location := *locationPtr
	_, err := os.Stat(location)
	if os.IsNotExist(err) {
		if strings.HasSuffix(location, "\\") != true {
			location = location + "\\"
		}
		os.MkdirAll(location, os.ModeDir)
	}
	bps := GetBlueprint(service, version)
	records := [][]string{
		{"Blueprint", "Priority", "Link", "Catagory", "Tag"},
	}
	for _, bp := range bps {
		rec := []string{bp.Title, bp.Priority, bp.Url, "新特性", bp.Series}
		records = append(records, rec)
	}

	// Write csv to std
	// w := csv.NewWriter(os.Stdout)
	// w.WriteAll(records)

	// Write csv to file
	fileName := getFileName(service, version, location)
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	fw := csv.NewWriter(f)
	fw.WriteAll(records)
	if err := fw.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	fmt.Println("OK, download complete: " + fileName)
}
