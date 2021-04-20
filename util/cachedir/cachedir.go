package cachedir

import (
	"os"
	"path"

	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

var cacheDirName = "fazz-ecr"

func getCacheFilePath(filename string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(cacheDir, cacheDirName, filename), nil
}

func LoadJSONFile(filename string, v interface{}) error {
	fullfilename, err := getCacheFilePath(filename)
	if err != nil {
		return err
	}

	return jsonfile.Read(fullfilename, v)
}

func SaveJSONFile(filename string, v interface{}) error {
	fullfilename, err := getCacheFilePath(filename)
	if err != nil {
		return err
	}

	return jsonfile.Write(fullfilename, v)
}
