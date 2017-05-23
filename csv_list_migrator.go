package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/financial-times/email-news-api/newsapi"
	"github.com/financial-times/email-platform-tools/config"
)

var limit int = 5

type PostBody struct {
	List string `json:"list"`
}

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
				u := make([]map[string]interface{}, 0)
				defer func() {
					wg.Done()
					<-sem
				}()
				body := PostBody{List: "55c8861dfdf6f00300b9f89a"}
				b, err := json.Marshal(body)
				if err == nil {
					_, err := c.PostURL("https://email-webservices.ft.com/users/"+rec+"/lists", b, &u)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("cool")
				}
			}(record[0])
		}
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return nil
}

func main() {
	var cfg config.Config
	if err := config.Bind("config_dev.yaml", &cfg); err != nil {
		fmt.Println(err)
	}
	r := &http.Client{}
	h := map[string]string{"Authorization": cfg.UsersAuth, "Content-Type": "application/json"}
	c := newsapi.NewClient(h, r)
	importUserIDs("mapping.csv", c)
}
