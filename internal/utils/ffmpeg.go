package utils

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/teemukurki/rip-tool/internal/common"
)

func handleTrackMap(tracks []int, trackType string) []string {
	if tracks == nil {
		return []string{
			"-map", fmt.Sprintf("0:%s?", trackType),
		}
	}
	var params []string
	for _, track := range tracks {
		params = append(params, "-map", fmt.Sprintf("0:%s:%d?", trackType, track))
	}
	return params
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

func FFmpegCmd(opts common.Options, input string, out string, trackLenght float32, additionalParams []string) *exec.Cmd {
	inputArgs := []string{
		"-analyzeduration", "300000000",
		"-probesize", "500M",
		"-i", input,
	}
	outputArgs := []string{
		"-c:a", opts.AudioCodec,
		"-c:s", "copy",
	}
	args := slices.Concat(
		inputArgs,
		additionalParams,
		handleTrackMap(opts.AudioTrack, "a"),
		handleTrackMap(opts.SubtitleTrack, "s"),
		handleTrackMap(opts.VideoTrack, "v"),
		outputArgs,
	)

	videoArgs := strings.Fields(opts.VideoEncodingParams)
	args = append(args, videoArgs...)
	if !opts.NoAutoLength {
		args = append(args, "-to", fmt.Sprintf("%.2fs", trackLenght))
	}
	args = append(args, out)

	fmt.Println("FFMPEG arguments ", args)
	return exec.Command("ffmpeg", args...)
}
