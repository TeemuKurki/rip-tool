package common

import (
	"os"
	"path/filepath"
)

type Options struct {
	Show                bool
	DiskPath            string
	Season              int
	Disk                int
	Titles              []int
	AudioCodec          string
	VerifySpeed         int
	MinOutputSize       int
	VideoEncodingParams string
	MinLength           int
	MaxLength           int
	AudioTrack          int
	AudioLang           string
	VideoTrack          int
	VideoLang           string
	SubtitleTrack       int
	SubtitleLang        string
	NoAutoLength        bool
	KeyPath             string
}

func home() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func DefaultOptions() Options {
	return Options{
		DiskPath:            "/dev/sr0",
		Season:              1,
		Disk:                1,
		AudioCodec:          "aac",
		VerifySpeed:         16,
		MinOutputSize:       10,
		VideoEncodingParams: "-c:v h264_nvenc -preset p7 -rc vbr -cq 23",
		MinLength:           20,
		MaxLength:           0,
		AudioTrack:          -1,
		SubtitleTrack:       -1,
		VideoTrack:          -1,
		NoAutoLength:        false,
		KeyPath:             filepath.Join(home(), ".config", "aacs", "KEYDB.cfg"),
	}
}
