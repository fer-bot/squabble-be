package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetRandomWord(length int) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://random-word-api.herokuapp.com/word?length=5", nil)
	if err != nil {
		fmt.Print(err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseObject []string
	json.Unmarshal(bodyBytes, &responseObject)
	return responseObject[0], nil
}
