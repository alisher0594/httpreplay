package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/alisher0594/httpreplay/fetcher"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Fetcher interface {
	Get() ([]fetcher.LoggedRequest, error)
}

func main() {
	req := flag.String("request", "", "Run a specific request")
	fr := flag.Bool("reqid", false, "Import failed requests from redis")
	flag.Parse()

	if len(*req) > 0 {
		run(*req)
		return
	}

	if *fr == false {
		flag.Usage()
		return
	}

	redis := *fr

	if exists, err := pathExists("./reqs"); err != nil {
		log.Fatal("unable to get directory stats")
	} else if exists == false {
		os.Mkdir("./reqs", 0777)
	}

	var f Fetcher
	if redis {
		f = fetcher.Redis{}
	}

	results, err := f.Get()
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range results {
		fn := fmt.Sprintf("./reqs/%s.log", r.ReqID)
		if err := ioutil.WriteFile(fn, r.Body, 0666); err != nil {
			log.Println("unable to write file", fn, err)
		}
	}
}

func run(req string) {
	f, err := os.Open(fmt.Sprintf("./reqs/%s.log", req))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	r, err := http.ReadRequest(buf)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode, r.Method, r.URL)
	fmt.Println(string(b))
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
