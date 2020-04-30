package main

import (
	"compress/bzip2"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/chrikoch/go-fah-alerter/config"
)

func main() {
	var configFilename string
	flag.StringVar(&configFilename, "config", "", "location of config-file")
	flag.Parse()

	if len(configFilename) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var c config.Config
	err := c.ReadFromFile(configFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(c)

	client := http.Client{}
	req, err := http.NewRequest("GET", "https://apps.foldingathome.org/daily_user_summary.txt.bz2", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("START getting user summary list\n")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("END getting list\n")
	log.Printf("response Header: %v\n", resp.Header)

	//TODO Etag und Last-Modified auswerten!!!

	bzip2Reader := bzip2.NewReader(resp.Body)

	csvReader := csv.NewReader(bzip2Reader)
	csvReader.Comma = '\t'
	csvReader.FieldsPerRecord = -1
	csvReader.ReuseRecord = true
	csvReader.LazyQuotes = true

	for {
		columns, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		} else if FindInSlice(c.UserNames, columns[0]) {
			fmt.Println(columns)
		}
	}
}

func FindInSlice(s []string, k string) bool {
	for _, i := range s {
		if i == k {
			return true
		}
	}

	return false
}
