package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func getUpdaterUri() string {
	return getAttendanceUrl("file_snapshots")
}

func getFileData(path string) ([]byte, error) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"relative_path": path,
		"contents":      string(fileData),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func postFile(path string) error {
	data, err := getFileData(path)
	if err != nil {
		return err
	}

	postData := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", getUpdaterUri(), postData)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Length", strconv.Itoa(len(data)))

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	response, err := client.Do(request)

	if err != nil {
		return err
	}
	log.Println("Updater Status:", response.Status)
	return nil
}

func SendFileUpdate(path string) error {
	log.Println("Uploading:", path)
	if len(getAttendanceId()) == 0 {
		return errors.New("no attendance id set; file update not sent")
	}
	return postFile(path)
}
