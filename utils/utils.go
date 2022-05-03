package utils

import (
	"errors"
	"log"
	"reflect"
)

// HandleErrors -> just logging the error at the moment
func HandleError(err error) {
	if err != nil {
		log.Println("An error occurred", err)
	}
}

// SplitToChunks -> splits slice to n number of slices
func SplitToChunks(slice []string, chunkNum int) [][]string {
	var chunked [][]string
	chunkSize := (len(slice) + chunkNum - 1) / chunkNum
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunked = append(chunked, slice[i:end])
	}
	return chunked
}

// FlattenDepthString -> flattens nested slices structure to a single slice
func FlattenDepthString(a reflect.Value, depth int) ([]string, error) {
	if depth < 1 {
		return []string{}, errors.New("input must be an slice of strings")
	} else if depth == 1 {
		stringArr := make([]string, a.Len())

		for i := 0; i < a.Len(); i++ {
			stringArr[i] = a.Index(i).String()
		}

		return stringArr, nil
	} else {
		stringArr := []string{}

		for i := 0; i < a.Len(); i++ {
			res, err := FlattenDepthString(a.Index(i), depth-1)
			if err != nil {
				return []string{}, err
			}
			stringArr = append(stringArr, res...)
		}
		return stringArr, nil
	}
}

// RemoveDuplicateStr -> removes duplicate values from a slice
func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
