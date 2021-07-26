package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const ASCII_ART = "  _____ _    _ _ _            __          ___           _\n" +
	" / ____| |  (_) | |           \\ \\        / / |         | |\n" +
	"| (___ | | ___| | | ___ _ __   \\ \\  /\\  / /| |__   __ _| | ___\n" +
	" \\___ \\| |/ / | | |/ _ \\ '__|   \\ \\/  \\/ / | '_ \\ / _` | |/ _ \\\n" +
	" ____) |   <| | | |  __/ |       \\  /\\  /  | | | | (_| | |  __/\n" +
	"|_____/|_|\\_\\_|_|_|\\___|_|        \\/  \\/   |_| |_|\\__,_|_|\\___|\n"

const DEFAULT_SERVER = "https://train.skillerwhale.com"

func Start() {
	var wg sync.WaitGroup
	fmt.Println(ASCII_ART)

	fmt.Println("We're going to start watching this directory for changes " +
		"so that the trainer can see your progress.")
	fmt.Println("Hit Ctrl+C to stop.")

	watcher := NewWatcher(getWatcherBasePath())

	wg.Add(1)
	go StartPing(time.Second, &wg)

	watcher.PollForChanges(time.Second)
	wg.Wait()
}

func getServerUrl() string {
	serverUrl, ok := os.LookupEnv("SERVER_URL")
	if !ok {
		return DEFAULT_SERVER
	}
	return serverUrl
}

func getAttendanceUrl(path string) string {
	attendanceId := getAttendanceId()
	fullPath := fmt.Sprintf(
		"attendances/%s/%s",
		url.QueryEscape(attendanceId),
		url.QueryEscape(path),
	)

	var serverUrl, err = url.Parse(getServerUrl())
	if err != nil {
		return ""
	}
	serverUrl.Path = fullPath
	return serverUrl.String()
}

func readAttendanceIdFile() string {
	attendanceIdFile, ok := os.LookupEnv("ATTENDANCE_ID_FILE")
	if !ok {
		log.Println("ATTENDANCE_ID_FILE environment variable not found")
		return ""
	}

	attendanceId, err := ioutil.ReadFile(attendanceIdFile)
	if err != nil {
		log.Println(err)
		return ""
	}

	return strings.TrimSpace(string(attendanceId))
}

func getAttendanceId() string {
	attendanceId, ok := os.LookupEnv("ATTENDANCE_ID")
	if ok {
		return attendanceId
	}
	return readAttendanceIdFile()
}

func getWatcherBasePath() string {
	basePath, ok := os.LookupEnv("WATCHER_BASE_PATH")
	if !ok {
		return "."
	}
	return basePath
}

func getWatchedExts() []string {
	return unmarshalJsonEnvVal("WATCHED_EXTS")
}

func getIgnoredDirs() []string {
	return unmarshalJsonEnvVal("IGNORE_DIRS")
}

func unmarshalJsonEnvVal(envVar string) []string {
	val, ok := os.LookupEnv(envVar)

	if !ok {
		return []string{}
	}
	var stringSlice []string
	err := json.Unmarshal([]byte(val), &stringSlice)

	if err != nil {
		log.Println(err)
		return []string{}
	}
	return stringSlice
}
