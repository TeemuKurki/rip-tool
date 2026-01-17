package bluray

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/teemukurki/rip-tool/internal/bluray/chapters"
	"github.com/teemukurki/rip-tool/internal/common"
	"github.com/teemukurki/rip-tool/internal/utils"
)

func toMinutes(time float32) float32 {
	return time / 60
}

type MatchLanguageTrack struct {
	LangCode    string
	Language    string
	StreamIndex int
}

func generateChapterMeta(c chapters.Chapter) []byte {
	result := ";FFMETADATA1\n\n"
	var endTime int = 0
	var startTime int = 0

	var chapters = c.Chapters

	for _, chapter := range chapters {
		startTime = endTime
		endTime += chapter.Duration

		result += "[CHAPTER]\n"
		result += "TIMEBASE=1/100000\n"
		result += fmt.Sprintf("START=%d\n", startTime)
		result += fmt.Sprintf("END=%d\n", endTime)
		result += fmt.Sprintf("title=Chapter \\#%d\n", chapter.ID)
		result += "\n"
	}
	return []byte(result)
}

func ripBluray(opts common.Options, title string, track BDTitle, chapters []chapters.Chapter) error {
	outputDirPath := utils.OutputPath(opts.Show, title, opts.Season)
	utils.CreateDir(outputDirPath)
	outputPath := filepath.Join(outputDirPath, fmt.Sprintf("%s-D%s-t%s.mkv", title, strconv.Itoa(opts.Disk), strconv.Itoa(track.Index)))

	if !utils.PromptFileDeletion(outputPath) {
		return nil
	}

	bdSlice := BdSpliceCmd(opts, track)
	langPassCmd := utils.FFmpegLangMetaCmd("-", track.PGLang, track.AudioLang)

	r1, w1, err := os.Pipe()
	if err != nil {
		fmt.Println("Error with os.Pipe", err)
		return err
	}
	defer r1.Close()

	// Prepare chapter information input pipeing
	additionalParams := []string{"-i", "pipe:3", "-map_metadata", "1"}
	ffmpeg := utils.FFmpegCmd(opts, "pipe:0", outputPath, float32(track.Duration), additionalParams)

	// Pipe chapter metadata into "pipe:3" input
	ffmpeg.ExtraFiles = []*os.File{r1}

	// Create pipe to stream bdSplice output to ffmpeg
	bdSplicePipe, err := bdSlice.StdoutPipe()
	if err != nil {
		fmt.Println("Error with pipeing bdSplice", err)
		return err
	}
	langPassPipe, err := langPassCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error with pipeing langPass", err)
		return err
	}

	// Pipe outputs. bdSplice -> langPass -> ffmpeg
	langPassCmd.Stdin = bdSplicePipe
	ffmpeg.Stdin = langPassPipe
	ffmpeg.Stderr = os.Stderr

	if err := bdSlice.Start(); err != nil {
		fmt.Println("Error with bdSlice start", err)
		return err
	}
	defer utils.TerminateProcess(bdSlice, 5*time.Second)

	if err := langPassCmd.Start(); err != nil {
		fmt.Println("Error with langPassCmd start", err)
		return err
	}
	defer utils.TerminateProcess(bdSlice, 5*time.Second)

	if err := ffmpeg.Start(); err != nil {
		fmt.Println("Error with ffmpeg start", err)
		return err
	}
	defer utils.TerminateProcess(ffmpeg, 5*time.Second)

	// Write to chapter pipe in goroutine after starting ffmpeg cmd
	chapter := chapters[track.Index-1]
	go func() {
		defer w1.Close() // Close after writing to signal EOF
		_, err := w1.Write(generateChapterMeta(chapter))
		if err != nil {
			fmt.Println("Error writing chapter metadata:", err)
		}
	}()

	bdSliceErr := bdSlice.Wait()
	if bdSliceErr != nil {
		fmt.Println("Error with bdSlice wait", bdSliceErr)
	}
	langPassErr := langPassCmd.Wait()
	if langPassErr != nil {
		fmt.Println("Error with langPass wait", langPassErr)
	}
	// ffmpeg might close before mvp if data stream is longer than track lenght in dvd metadata.
	// In this case mvp will throw error, close itself and terminateProcess kills proceess PID
	// Video ripping should be successfull in this scenario
	ffmpegWaitErr := ffmpeg.Wait()
	if ffmpegWaitErr != nil {
		fmt.Println("Error with ffmpeg wait", ffmpegWaitErr)
	}

	return nil
}

func RunBluray(opts common.Options, title string) error {
	tracks, err := GetDBTitles(opts)
	if err != nil {
		fmt.Println("Failed to get db_list_titles data", err)
	}
	chapters, err := chapters.GetChapters()
	if err != nil {
		fmt.Println("Failed to get dchapter_list_titles data", err)
	}

	if len(opts.Titles) > 0 {
		for _, trackTitle := range opts.Titles {
			trackI := slices.IndexFunc(tracks, func(t BDTitle) bool {
				// Titles start from zero, Indexes start from 1
				return t.Index == trackTitle+1
			})
			if trackI > -1 {
				fmt.Printf("Ripping title %d\n", trackTitle)
				fmt.Printf("%+v\n", opts)
				track := tracks[trackI]
				fmt.Printf("%+v\n", track)
				ripBluray(opts, title, track, chapters)
			}

		}
	} else {
		for _, track := range tracks {
			if toMinutes(float32(track.Duration)) >= float32(opts.MinLength) && (opts.MaxLength == 0 || toMinutes(float32(track.Duration)) <= float32(opts.MaxLength)) {
				fmt.Printf("%+v\n", track)
				ripBluray(opts, title, track, chapters)
			}
		}
	}

	return nil
}
