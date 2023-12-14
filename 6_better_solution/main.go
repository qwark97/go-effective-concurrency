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
	defer log.Printf("(worker) INFO: end of work")

	var (
		wg sync.WaitGroup
	)
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	dataCh, errCh := streamer(ctx)
	go func() {
		defer wg.Done()
		for err := range errCh {
			log.Printf("(worker) ERR: err from stream: %v", err)
		}
	}()
	defer drain(dataCh)
	defer cancel()

	i := 0
	for data := range dataCh {
		name, err := process(data)
		if err != nil {
			return err
		}
		log.Printf("(worker) INFO: Fantastic name %d: %s", i, name)
		i++
	}
	wg.Wait()
	return nil
}

func drain[T any](channel <-chan T) {
	for range channel {
	}
}

func streamer(ctx context.Context) (chan Data, chan error) {
	var (
		dataCh = make(chan Data)
		errCh  = make(chan error)
	)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		for idx, d := range dataSets {
			if ctx.Err() != nil {
				break
			}
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

func process(d Data) (string, error) {
	if d.Name == "Rachel" {
		return "", fmt.Errorf("gopher can't stand this person")
	}
	return d.Name, nil
}
