package cachedir

import (
	"os"
	"path/filepath"

	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

var cacheDirName = "fazz-ecr"

func getCacheFilePath(filename string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err)
	}

	dir := filepath.Join(cacheDir, cacheDirName)
	os.MkdirAll(dir, 0o700)

	return filepath.Join(dir, filename), nil
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
