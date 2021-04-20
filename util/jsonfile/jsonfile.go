package jsonfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/payfazz/go-errors"
)

func Read(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func Write(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err)
	}

	tempFileName := fmt.Sprintf("%s-%d-%d", filename, time.Now().UnixNano(), uint8(rand.Int()))
	if err := ioutil.WriteFile(tempFileName, data, 0o600); err != nil {
		os.Remove(tempFileName)
		return err
	}

	return os.Rename(tempFileName, filename)
}
