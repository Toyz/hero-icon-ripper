package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/shettyh/threadpool"
)

const (
	baseURL = "https://d1u1mce87gyfbn.cloudfront.net/game/heroes/small/%s"
)

var (
	start  int
	stop   int
	folder string
)

func init() {
	flag.IntVar(&start, "start", 0, "Start amount for loop")
	flag.IntVar(&stop, "stop", 999, "Stop amount for loop")
	flag.StringVar(&folder, "folder", "./", "Folder to save files into defaults to ./")
}

func main() {
	pool := threadpool.NewThreadPool(10, 1000000)
	flag.Parse()

	_ = os.Mkdir(folder, 0777)

	var waitgroup sync.WaitGroup
	for i := start; i < stop; i++ {
		h := fmt.Sprintf("0x02E%013X.png", i)

		log.Printf("Downloading icon: %s", h)

		waitgroup.Add(1)
		pool.Execute(&worker{
			url:          fmt.Sprintf(baseURL, h),
			id:           h,
			saveLocation: path.Join(folder, h),
			waitgroup:    &waitgroup,
		})
	}

	waitgroup.Wait()
	pool.Close()
}

type worker struct {
	url          string
	id           string
	saveLocation string
	waitgroup    *sync.WaitGroup
}

func (t *worker) Run() {
	// Do your task here
	if err := t.downloadFile(t.saveLocation, t.url); err != nil {
		log.Println(err)
	} else {
		log.Printf("Downloaded: %s", t.url)
	}

	t.waitgroup.Done()
}

func (t *worker) downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		return err
	}
	return fmt.Errorf("Failed to download: %s", url)
}
