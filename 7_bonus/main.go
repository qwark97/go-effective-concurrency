package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

var dataSets = [...]string{
	`{"name": "Mark"}`,
	`{"name": "Alice"}`,
	`{"name": "Rachel"}`,
	`{"name": "Ella"}`,
	`{"name": "David"}`,
}

type Data struct {
	Name string `json:"name"`
}

func main() {
	if err := worker(); err != nil {
		log.Println("finished main unhappily, reason: " + err.Error())
	} else {
		log.Println("finished main successfully")
	}

	// select{} <- doesn't work anymore (check since when)
	// <-make(chan struct{}) <- doesn't work anymore (check since when)
	// for { <- works, but it is bad idea
	// }
	time.Sleep(1<<63 - 1)
}

func worker() error {
	log.Printf("(worker) INFO: start of work")
	defer log.Printf("(worker) INFO: end of work")

	var (
		wg sync.WaitGroup
	)
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dataCh, errCh := streamer()
	go func() {
		defer wg.Done()
		for err := range errCh {
			log.Printf("(worker) ERR: err from stream: %v", err)
		}
	}()

	for n := 0; n < 3; n++ {
		go processor(ctx, n, dataCh)
	}

	wg.Wait()
	return nil
}

func streamer() (chan Data, <-chan error) {
	var (
		dataCh = make(chan Data)
		errCh  = make(chan error)
	)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		for idx, d := range dataSets {
			var container Data
			err := json.Unmarshal([]byte(d), &container)
			if err != nil {
				errCh <- err
				continue
			}

			log.Printf("(streamer) INFO: waiting to send idx %d", idx)
			dataCh <- container
			log.Printf("(streamer) INFO: sent idx %d", idx)
		}
		log.Printf("(streamer) INFO: finished streaming <---- THE GOAL")
	}()

	return dataCh, errCh
}

func processor(ctx context.Context, id int, inputCh chan Data) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-inputCh: // when reading like tha, last value will be empty
			if !ok {
				return
			}
			if data.Name == "Rachel" {
				err := fmt.Errorf("gopher can't stand this person")
				log.Printf("(processor #%d) ERR: %v", id, err)
				continue
			}
			log.Printf("(processor #%d) INFO: Fantastic name %s", id, data.Name)
		}
	}
}
