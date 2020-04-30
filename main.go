package main

import (
	"compress/bzip2"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chrikoch/go-fah-alerter/config"

	pushb "github.com/xconstruct/go-pushbullet"
)

type FaHState struct {
	eTag         string
	lastModified string
}

type FaHChecker struct {
	state     FaHState
	usernames config.UserNameList
}

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
		log.Println(err)
		os.Exit(1)
	}

	log.Println(c)

	pc := pushb.New(c.Pushbullet.APIkey)
	err = pc.PushNote(c.Pushbullet.DeviceIdent, "hallo", "welt")
	if err != nil {
		log.Println(err)
	}

	checker := FaHChecker{usernames: c.UserNames}

	for {
		checker.CheckForNewUserData()
		time.Sleep(time.Second * 1800)
	}

}

func (f *FaHChecker) CheckForNewUserData() {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://apps.foldingathome.org/daily_user_summary.txt.bz2", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("If-None-Match", f.state.eTag)
	req.Header.Add("If-Modified-Since", f.state.lastModified)

	log.Printf("START getting user summary list\n")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("END getting list\n")
	log.Printf("response Header: %v\n", resp.Header)
	log.Printf("HTTP-Status: %v\n", resp.StatusCode)

	f.state.eTag = resp.Header.Get("Etag")
	f.state.lastModified = resp.Header.Get("Last-Modified")

	if resp.StatusCode != 200 {
		log.Println("Status != 200, nothing to do.")
		return
	}
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
			log.Println(err)
		} else if FindInSlice(f.usernames, columns[0]) {
			log.Println(columns)
		}
	}
}

//FindInSlice returns true, if k is found in s
func FindInSlice(s []string, k string) bool {
	for _, i := range s {
		if i == k {
			return true
		}
	}

	return false
}
