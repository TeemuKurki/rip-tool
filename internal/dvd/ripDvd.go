package dvd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/teemukurki/rip-tool/internal/common"
	"github.com/teemukurki/rip-tool/internal/utils"
)

type MatchLanguageTrack struct {
	LangCode    string
	Language    string
	StreamIndex int
}

func update_track_lang(opts common.Options, videoFilePath string, subtitleInfos []Subp, audioInfos []Audio, ffmpegStreamsInfo []utils.StreamInfo) {
	var matches []MatchLanguageTrack

	for _, stream := range ffmpegStreamsInfo {
		subtitleMatchIndex := slices.IndexFunc(subtitleInfos, func(subtitleInfo Subp) bool {
			return stream.HexID == subtitleInfo.StreamID
		})
		if subtitleMatchIndex > -1 {
			info := subtitleInfos[subtitleMatchIndex]
			matches = append(matches, MatchLanguageTrack{
				LangCode:    info.LangCode,
				Language:    info.Language,
				StreamIndex: stream.StreamIndex,
			})
		} else {
			audioMatchIndex := slices.IndexFunc(audioInfos, func(audioInfo Audio) bool {
				return stream.HexID == audioInfo.StreamID
			})
			if audioMatchIndex > -1 {
				info := audioInfos[audioMatchIndex]
				matches = append(matches, MatchLanguageTrack{
					LangCode:    info.LangCode,
					Language:    info.Language,
					StreamIndex: stream.StreamIndex,
				})
			}
		}
	}

	for _, match := range matches {
		args := []string{
			videoFilePath,
			"--edit", fmt.Sprintf("track:%d", match.StreamIndex),
			"--set", fmt.Sprintf("language=%s", match.LangCode),
			"--set", fmt.Sprintf("name=%s", match.Language),
		}

		if opts.AudioLang != "" && (opts.AudioLang == match.LangCode || opts.AudioLang == match.Language) {
			args = append(args, "--set", "flag-default=1")
		}
		if opts.VideoLang != "" && (opts.VideoLang == match.LangCode || opts.VideoLang == match.Language) {
			args = append(args, "--set", "flag-default=1")
		}
		if opts.SubtitleLang != "" && (opts.SubtitleLang == match.LangCode || opts.SubtitleLang == match.Language) {
			args = append(args, "--set", "flag-default=1")
		}

		cmd := exec.Command("mkvpropedit", args...)

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("mkvpropedit failed: %v\nOutput: %s\n", err, out)
		}
	}
}

func ripDVD(opts common.Options, trackId int, title string, subtitleInfo []Subp, audioInfo []Audio, trackLength float32) error {
	tmpDir, _ := os.MkdirTemp("", "rip-stream")
	defer os.RemoveAll(tmpDir)

	fifo := filepath.Join(tmpDir, "stream.fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		return fmt.Errorf("mkfifo failed: %w", err)
	}

	outputDirPath := utils.OutputPath(opts.Show, title, opts.Season)

	utils.CreateDir(outputDirPath)

	outputPath := filepath.Join(outputDirPath, fmt.Sprintf("%s-D%s-t%s.mkv", title, strconv.Itoa(opts.Disk), strconv.Itoa(trackId)))

	if utils.PromptFileDeletion(outputPath) == false {
		return nil
	}

	//TODO: Look into using exec.CommandContext() instead
	mpv := exec.Command(
		"mpv",
		fmt.Sprintf("dvd://%d", trackId),
		"--dvd-device="+opts.DiskPath,
		"--stream-dump="+fifo,
	)

	mpv.Stderr = os.Stderr // Pass mpv log/progress to terminal

	ffmpeg := utils.FFmpegCmd(opts, fifo, outputPath, trackLength)

	stderrPipe, err := ffmpeg.StderrPipe()
	if err != nil {
		fmt.Println("Error with ffmpeg StderrPipe", err)

		return err
	}
	tee := io.TeeReader(stderrPipe, os.Stderr) // Pass ffmpeg log/progress to terminal, also tee logs for subtitle/audio parsing

	if err := mpv.Start(); err != nil {
		fmt.Println("Error with mpv start", err)

		return err
	}
	fmt.Printf("Start dumpting data to %s for processing", fifo)
	defer utils.TerminateProcess(mpv, 5*time.Second)
	mvpPid := mpv.Process.Pid
	fmt.Println("mpv PID:", mvpPid)

	if err := ffmpeg.Start(); err != nil {
		fmt.Println("Error with ffmpeg start", err)
		return err
	}

	ffmpegPid := ffmpeg.Process.Pid
	fmt.Println("ffmpeg PID:", ffmpegPid)
	defer utils.TerminateProcess(ffmpeg, 5*time.Second)

	streams, parseStreamErr := utils.ParseStreams(tee, os.Stdout)
	if parseStreamErr != nil {
		fmt.Println("Error with ParseStreams", parseStreamErr)
		return err
	}

	mvpWaitErr := mpv.Wait()
	if mvpWaitErr != nil {
		fmt.Println("Error with mvp wait", mvpWaitErr)
	}
	// ffmpeg might close before mvp if data stream is longer than track lenght in dvd metadata.
	// In this case mvp will throw error, close itself and terminateProcess kills proceess PID
	// Video ripping should be successfull in this scenario
	ffmpegWaitErr := ffmpeg.Wait()
	if ffmpegWaitErr != nil {
		fmt.Println("Error with ffmpeg wait", ffmpegWaitErr)
	}

	fmt.Println("Streams found:", streams)
	update_track_lang(opts, outputPath, subtitleInfo, audioInfo, streams)
	return nil
}

func toMinutes(time float32) float32 {
	return time / 60
}

func RunDVD(opts common.Options, title string) error {
	lsdvd, err := GetLsdvdInfo(opts)
	if err != nil {
		fmt.Println("Failed to get lsdvd data", err)
	}

	var tracks []Track
	if lsdvd != nil {
		tracks = lsdvd.Tracks
	}

	if len(opts.Titles) > 0 {
		for _, track := range opts.Titles {
			trackI := slices.IndexFunc(tracks, func(t Track) bool {
				return t.Index == track
			})
			if trackI > -1 {
				fmt.Printf("Ripping title %d\n", track)
				fmt.Printf("%+v\n", opts)
				track := tracks[trackI]
				ripDVD(opts, track.Index-1, title, track.Subps, track.Audio, track.Length)
			}

		}

	} else {
		for _, track := range tracks {
			if toMinutes(track.Length) >= float32(opts.MinLength) && (opts.MaxLength == 0 || toMinutes(track.Length) <= float32(opts.MaxLength)) {
				fmt.Printf("%+v\n", opts)
				ripDVD(opts, track.Index-1, title, track.Subps, track.Audio, track.Length)
			}
		}
	}

	return nil
}
