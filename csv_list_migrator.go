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
	"sync/atomic"

	"github.com/financial-times/email-news-api/newsapi"
	"github.com/financial-times/email-platform-tools/config"
)

var limit int = 20

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
	var count uint64 = 0
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		countFinal := atomic.LoadUint64(&count)
		fmt.Println("Final:", countFinal)
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
				body := PostBody{List: "58db721900eb6f0004d56a23"}
				b, err := json.Marshal(body)
				if err == nil {
					_, err := c.PostURL("https://email-webservices.ft.com/users/"+rec+"/lists", b, &u)
					if err != nil {
						fmt.Println(err)
					}
					atomic.AddUint64(&count, 1)
					fmt.Println(count)
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
	importUserIDs("users.csv", c)
}
