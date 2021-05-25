package jsonfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/util/randstring"
)

func Read(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Trace(err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func Write(filename string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		return errors.Trace(err)
	}

	data, err := json.MarshalIndent(v, "", "	")
	if err != nil {
		return errors.Trace(err)
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	tempFileName := fmt.Sprintf("%s-%s", filename, randstring.Get(10))
	defer os.Remove(tempFileName)

	if err := ioutil.WriteFile(tempFileName, data, 0o600); err != nil {
		return errors.Trace(err)
	}

	if err := os.Rename(tempFileName, filename); err != nil {
		return errors.Trace(err)
	}

	return nil
}
