package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/teemukurki/rip-tool/internal/common"
)

func handleTrackMap(track int) string {
	if track == -1 {
		return ""
	}
	return fmt.Sprintf(":%d", track)
}

func genLangMetadata(langData []string, metaType string) []string {
	var result []string

	for i, lang := range langData {
		result = append(result, fmt.Sprintf("-metadata:s:%s:%d", metaType, i), fmt.Sprintf("language=%s", lang))
	}
	return result
}

// Insert audio and subtitle language metadata to stream
func FFmpegLangMetaCmd(input string, subtitleLangs []string, audioLangs []string) *exec.Cmd {
	args := []string{
		"-analyzeduration", "300000000",
		"-probesize", "500M",
		"-i", input,
		"-hide_banner",
		"-map", "0",
		"-c", "copy",
		"-f", "matroska",
	}
	args = append(args, genLangMetadata(subtitleLangs, "s")...)
	args = append(args, genLangMetadata(audioLangs, "a")...)

	args = append(args, "-")

	return exec.Command("ffmpeg", args...)
}

func FFmpegCmd(opts common.Options, input string, out string, trackLenght float32) *exec.Cmd {

	args := []string{
		"-analyzeduration", "300000000",
		"-probesize", "500M",
		"-i", input,
		"-map", fmt.Sprintf("0:a%s?", handleTrackMap(opts.AudioTrack)),
		"-map", fmt.Sprintf("0:s%s?", handleTrackMap(opts.SubtitleTrack)),
		"-map", fmt.Sprintf("0:v%s", handleTrackMap(opts.VideoTrack)),
		"-c:a", opts.AudioCodec,
		"-c:s", "copy",
		out,
	}
	videoArgs := strings.Fields(opts.VideoEncodingParams)
	args = append(args, videoArgs...)
	if !opts.NoAutoLength {
		args = append(args, "-to", fmt.Sprintf("%.2fs", trackLenght))
	}

	fmt.Println("FFMPEG arguments ", args)
	return exec.Command("ffmpeg", args...)
}
