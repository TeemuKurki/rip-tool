package chapters

/*
#cgo pkg-config: libbluray
// Portions of the C code included here are adapted from libbluray/tools/db_list_titles.c
// (C) 2009-2010 John Stebbins, LGPL v2.1+
#include "chapters.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type ChapterInfo struct {
	ID       int `json:"id"`
	Duration int `json:"duration"`
	Title    int `json:"title"`
}

type Chapter struct {
	Title    int
	Chapters []ChapterInfo
}

func GetChapters() ([]Chapter, error) {
	path := C.CString("/dev/sr0")
	defer C.free(unsafe.Pointer(path))

	jsonC := C.run_chapters(path)
	defer C.free(unsafe.Pointer(jsonC))

	jsonGo := C.GoString(jsonC)

	if jsonC == nil {
		return nil, fmt.Errorf("Chapters returned NULL. something went wrong!")
	}

	var chaptersInfos [][]ChapterInfo
	if err := json.Unmarshal([]byte(jsonGo), &chaptersInfos); err != nil {
		return nil, fmt.Errorf("Unmashaling JSON string into chapters failed. %w", err)
	}

	var chapters []Chapter
	for _, title := range chaptersInfos {
		chapters = append(chapters, Chapter{
			Title:    title[0].Title,
			Chapters: title,
		})
	}

	return chapters, nil
}
