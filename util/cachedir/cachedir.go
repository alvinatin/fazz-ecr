package cachedir

import (
	"os"
	"path/filepath"

	"github.com/payfazz/fazz-ecr/util/jsonfile"
	"github.com/payfazz/go-errors"
)

var cacheDirName = "fazz-ecr"

func getCacheFilePath(filename string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err)
	}

	return filepath.Join(cacheDir, cacheDirName, filename), nil
}

func LoadJSONFile(filename string, v interface{}) error {
	fullfilename, err := getCacheFilePath(filename)
	if err != nil {
		return errors.Wrap(err)
	}

	return jsonfile.Read(fullfilename, v)
}

func SaveJSONFile(filename string, v interface{}) error {
	fullfilename, err := getCacheFilePath(filename)
	if err != nil {
		return errors.Wrap(err)
	}

	return jsonfile.Write(fullfilename, v)
}
