package main

import (
	"fmt"
	"os"
	"time"

	"github.com/barasher/go-exiftool"
)

func main() {
	var keepTime string

	fmt.Println("Insert last date for photos to keep in format 2023/12/15:")
	fmt.Scanln(&keepTime)

	cutoffDate, err := time.Parse("2006/01/02", keepTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing your date, you probably put it in wrong: %s\n", err.Error())
		return
	}

	// TODO: (Matei) Could ask from terminal line later, maybe.
	photosDir := "photos/"

	// Save files for "deletion" and then place in other directory
	// for safety reasons. We'll also write down the "deleted" photos
	// in a text file.
	type DeletedPhoto struct {
		time time.Time
		name string
	}
	deletedPhotos := make([]DeletedPhoto, 0)

	files, err := os.ReadDir(photosDir)
	if err != nil {
		panic(err)
	}

	et, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when intializing: %v\n", err)
		return
	}
	defer et.Close()

	for _, file := range files {
		exifData := et.ExtractMetadata(photosDir + file.Name())

		var photoTimeField string
		for _, data := range exifData {
			if data.Err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %v: %v\n", data.File, data.Err)
				return
			}

			photoTimeField, err = data.GetString("CreateDate")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read CreateDate field: %s", err.Error())
				return
			}
		}

		photoTime, err := time.Parse("2006:01:02 15:04:05", photoTimeField)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse time for photo: %s\n", err.Error())
			return
		}

		if photoTime.After(cutoffDate) {
			deletedPhotos = append(deletedPhotos, DeletedPhoto{
				name: file.Name(),
				time: photoTime,
			})
		}
	}

	// Create directory for the deleted photos.
	err = os.Mkdir("deletedPhotos", os.FileMode(0755))
	if err != nil {
		panic(err)
	}

	// Create file to write deleted files in.
	helpFile, err := os.Create("README.txt")
	if err != nil {
		panic(err)
	}

	helpText := "I've set aside all of your files for deletion.\nJust delete the 'deletedPhotos' directory.\nHere's a list of the stuff I set aside for you, with dates before your cutoff date of " + keepTime + ":\n"

	for _, photo := range deletedPhotos {
		err := os.Rename("photos/"+photo.name, "deletedPhotos/"+photo.name)
		if err != nil {
			panic(err)
		}
		helpText += fmt.Sprintf("%s - Created on %s\n", photo.name, photo.time.Format(time.RFC1123))
	}

	_, err = helpFile.WriteString(helpText)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}
