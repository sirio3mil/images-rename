package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/tidwall/gjson"
)

const imagesFolderTo = "E:\\OneDrive\\Imágenes\\Álbum de cámara"
const imagesFolderFrom = "E:\\OneDrive\\Imágenes\\Overon"

func readFiles(root string) (bool, error) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		matched, _ := regexp.MatchString(`\\2\d{3}\\\d{2}\\`, path)
		if !matched {
			fileExtension := strings.ToLower(filepath.Ext(path))
			switch fileExtension {
			case ".jpg":
				log.Printf("%s\n", path)
				return moveJpeg(path, info)
			default:
				log.Printf("%s\n", path)
				return moveDefault(path, info)
			}
		}
		return nil
	})

	return true, nil
}

func moveJpeg(path string, info os.FileInfo) error {
	ok, err := moveJPEGFileWithExif(path)
	if ok {
		return nil
	}
	if err != nil {
		log.Println(err)
	}

	return moveDefault(path, info)
}

func moveDefault(path string, info os.FileInfo) error {
	ok, err := moveFileWithPath(path)
	if ok {
		return nil
	}
	if err != nil {
		log.Println(err)
	}
	ok, err = moveFileWithFileInfo(path, info)
	if ok {
		return nil
	}
	if err != nil {
		log.Println(err)
	}
	return nil
}

func moveFileWithFileInfo(path string, info os.FileInfo) (bool, error) {
	year, month := getYearMonthFromFileInfo(info)
	ok, err := moveFile(path, year, month)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

func moveFileWithPath(path string) (bool, error) {
	year, month, err := getYearMonthFromFilePath(path)
	if err != nil {
		return false, err
	}
	ok, err := moveFile(path, year, month)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

func getMetadata(fname string) (*exif.Exif, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metaData, err := exif.Decode(file)
	if err != nil {
		return nil, err
	}

	return metaData, nil
}

func moveJPEGFileWithExif(fname string) (bool, error) {
	metaData, err := getMetadata(fname)
	if err != nil {
		return false, err
	}
	year, month, err := getYearMonthFromMetadata(metaData)
	if err != nil {
		return false, err
	}
	ok, err := moveFile(fname, year, month)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

func getYearMonthFromFileInfo(info os.FileInfo) (string, string) {
	dateTime := info.ModTime()
	year := fmt.Sprintf("%d", dateTime.Year())
	month := fmt.Sprintf("%02d", dateTime.Month())

	return year, month
}

func getYearMonthFromFilePath(path string) (string, string, error) {
	_, fileName := filepath.Split(path)
	re := regexp.MustCompile(`(\d{4})(\d{2})(\d{2})`)
	matchs := re.FindStringSubmatch(fileName)
	if len(matchs) < 4 {
		return "", "", errors.New("Year and month not found in file path")
	}

	return matchs[1], matchs[2], nil
}

func getYearMonthFromMetadata(metaData *exif.Exif) (string, string, error) {
	jsonByte, err := metaData.MarshalJSON()
	if err != nil {
		return "", "", err
	}

	jsonString := string(jsonByte)
	dateTime := gjson.Get(jsonString, "DateTimeOriginal").String()
	if dateTime == "" {
		dateTime = gjson.Get(jsonString, "DateTimeDigitized").String()
		if dateTime == "" {
			dateTime = gjson.Get(jsonString, "DateTime").String()
			if dateTime == "" {
				return "", "", errors.New("DateTime tag not found in EXIF blob")
			}
		}
	}
	s := strings.Split(dateTime, " ")
	date := s[0]
	s = strings.Split(date, ":")

	return s[0], s[1], nil
}

func moveFile(fname string, year string, month string) (bool, error) {
	_, fileName := filepath.Split(fname)
	newDir := imagesFolderTo + "\\" + year + "\\" + month + "\\"
	_, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(newDir, 0777)
		if err != nil {
			return false, err
		}
	}
	newLocation := newDir + fileName
	fmt.Println(newLocation)
	err = os.Rename(fname, newLocation)
	if err != nil {
		return false, err
	}

	return true, nil
}

func main() {
	_, err := readFiles(imagesFolderFrom)
	if err != nil {
		log.Fatal(err)
	}
}
