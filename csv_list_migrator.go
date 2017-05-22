package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/financial-times/email-news-api/newsapi"
)

var limit int = 5

func importUserIDs(path string, c *newsapi.Client) error {
	sem := make(chan bool, limit)
	if _, err := os.Stat(path); err != nil {
		return errors.New("Config path not valid")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	bufr := bufio.NewReader(f)
	r := csv.NewReader(bufr)
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(sem)
	}()

	// read header
	_, err = r.Read()
	if err != nil {
		return err
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) > 0 {
			sem <- true
			wg.Add(1)
			go func(rec string) {
				u := make(map[string]interface{})
				defer func() {
					wg.Done()
					<-sem
				}()
				_, err := c.GetURL("https://email-webservices.ft.com/users/"+rec, &u)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(u)
			}(record[0])
		}
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return nil
}

func main() {
	r := &http.Client{}
	h := map[string]string{"Authorization": "Basic "}
	c := newsapi.NewClient(h, r)
	importUserIDs("mapping.csv", c)
}
