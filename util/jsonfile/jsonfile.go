package jsonfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

func ReadFile(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func WriteFile(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	tempFileName := fmt.Sprintf("%s-%d-%d", filename, time.Now().UnixNano(), uint8(rand.Int()))
	if err := ioutil.WriteFile(tempFileName, data, 0o644); err != nil {
		os.Remove(tempFileName)
		return err
	}

	return os.Rename(tempFileName, filename)
}
