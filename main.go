package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func main() {

	// credentials and task
	credentials := flag.String("credentials", "", "auth credentials")
	host := flag.String("host", "", "url")
	taskName := flag.String("task", "", "task name")
	attempts := flag.Int("attempts", 50, "number of attempts")
	flag.Parse()

	// validate credentials
	if credentials == nil || *credentials == "" {
		logrus.Errorf("credentials not provided")
		return
	}

	// validate task
	if taskName == nil || *taskName == "" {
		logrus.Errorf("task not provided")
		return
	}

	// validate host
	if host == nil || *host == "" {
		logrus.Errorf("host not provided")
		return
	}

	var headers = make(map[string]string)
	authCredentials, err := readFile(*credentials)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}
	headers["Authorization"] = fmt.Sprintf("Basic %s", strings.TrimSpace(authCredentials))

	taskRequest := TaskRequest{Name: *taskName}

	body, err := json.Marshal(taskRequest)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/task", *host), strings.NewReader(string(body)))
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Error: %v", err)
		return
	}

	// http request poll
	getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/task/%s", *host, *taskName), nil)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}

	for key, value := range headers {
		getReq.Header.Add(key, value)
	}

	for i := 0; i < *attempts; i++ {

		resp, err := client.Do(getReq)
		if err != nil {
			logrus.Errorf("Error: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			logrus.Errorf("Error: %v", err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("Error: %v", err)
			return
		}

		if strings.Contains(string(body), "COMPLETED") {
			fmt.Println("COMPLETED")
			return
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("FAILED")
}

func readFile(filePath string) (string, error) {
	var out, errOut bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s", filePath))
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("unable to run command, err: %v; error output: %v", err, errOut.String())
	}
	return out.String(), nil
}

type TaskRequest struct {
	Name string `json:"name"`
}
