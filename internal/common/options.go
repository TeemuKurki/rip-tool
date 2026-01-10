package common

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
