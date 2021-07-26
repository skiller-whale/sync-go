package sync

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"
)

func getPingUri() string {
	return getAttendanceUrl("pings")
}

func ping() error {
	request, err := http.NewRequest("POST", getPingUri(), bytes.NewBuffer([]byte{}))

	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	_, err = client.Do(request)

	return err
}

func StartPing(waitTime time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		err := ping()
		if err != nil {
			// We want to continue looping even if we hit an unexpected error
			log.Println("Unexpected error with ping:", err)
		}

		// Send a ping every `waitTime` seconds
		time.Sleep(waitTime)
	}
}
