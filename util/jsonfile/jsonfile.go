package jsonfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/util/randstring"
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

	tempFileName := fmt.Sprintf("%s-%s", filename, randstring.Get(10))
	if err := ioutil.WriteFile(tempFileName, data, 0o600); err != nil {
		os.Remove(tempFileName)
		return errors.Wrap(err)
	}

	if err := os.Rename(tempFileName, filename); err != nil {
		return errors.Wrap(err)
	}

	return nil
}
