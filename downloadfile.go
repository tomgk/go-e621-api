package e621api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type MaxLimitExceeded string

func (e MaxLimitExceeded) Error() string {
	return string(e)
}

type FileCache interface {
	//returns the cached file for the given URL
	//
	//if err is set an error occured and file and filename have unspecified valus
	//otherwise if file is set the file has yet to be downloaded into file
	//otherwise if only filename is set the content has already been downloaded and are in the file of that name
	ForUrl(url string) (file *os.File, filename string, err error)

	String() string
}

type TemporaryFileCache struct{}

func (c *TemporaryFileCache) ForUrl(url string) (file *os.File, filename string, err error) {
	file, err = ioutil.TempFile("", "something")
	return
}

func (c *TemporaryFileCache) String() string {
	return "TemporaryFileCache"
}

var _ FileCache = &TemporaryFileCache{}

type PathExtractor interface {
	GetPath(url string) (string, error)
}

type E621PathExtractor struct {
}

func (e *E621PathExtractor) GetPath(urlString string) (string, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	if url.Hostname() != "static1.e621.net" {
		return "", fmt.Errorf("Wrong host %v", url.Hostname())
	}

	path := strings.FieldsFunc(url.Path, func(c rune) bool {
		return c == '/'
	})

	// "/data" [...]
	if path[0] != "data" {
		return "", fmt.Errorf("Path not starting with data but %v", path[0])
	}

	path = path[1:]

	var prefix string

	// "ab/cd/abcd...extension
	// "preview/ab/cd/abcd...extension
	// "sample/ab/cd/abcd...extension
	if len(path) == 4 {
		prefix = path[0] + "_"
		path = path[1:]
	} else if len(path) != 3 {
		return "", fmt.Errorf("Path segment count not correct")
	}

	filename := path[len(path)-1]

	return prefix + filename, nil
}

var _ PathExtractor = &E621PathExtractor{}

type DirectoryFileCache struct {
	Directory     string
	PathExtractor PathExtractor
}

func (c *DirectoryFileCache) ForUrl(url string) (file *os.File, filename string, err error) {
	var path string
	path, err = c.PathExtractor.GetPath(url)
	if err != nil {
		return nil, "", err
	}

	filename = c.Directory + "/" + path

	//open with O_EXCL so it fails if file already exists
	file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755) //os.Create(c.directory+"/"+path)
	//already existing is not an error but means that the file was already downloaded
	if err != nil && errors.Is(err, os.ErrExist) {
		err = nil
	}

	return
}

func (c *DirectoryFileCache) String() string {
	//return fmt.Sprintf(`DirectoryFileCache(directory: "%s", pathExtractor:%v)`, c.directory, c.pathExtractor)
	return fmt.Sprintf(`DirectoryFileCache(directory: "%s")`, c.Directory)
}

var _ FileCache = &DirectoryFileCache{}

func DownloadFile(cache FileCache, url string) (filename string, error error) {
	defer func() {
		if error != nil {
			log.Printf("Downloading %v failed: %s\n", url, error)
		}
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		error = err
		return
	}
	defer resp.Body.Close()

	const maximum = 1024 * 1024 * 30

	//TODO: resp.ContentLength could be -1
	if resp.ContentLength > maximum {
		error = MaxLimitExceeded(fmt.Sprintf("Maximum size of 10 MB exceeded, has %v", resp.ContentLength/1024/1024))
		return
	}

	var file *os.File

	file, filename, error = cache.ForUrl(url)
	if err != nil {
		error = err
		return
	} else if file != nil {
		defer file.Close()

		log.Printf("Cache miss for %v, downloading...\n", url)

		filename = file.Name()

		// Write the body to file
		_, error = io.Copy(file, resp.Body)
		return
	} else {
		log.Printf("Cache hit for %v\n", url)
		return
	}
}
