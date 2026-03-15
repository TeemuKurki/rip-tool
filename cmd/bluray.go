package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/teemukurki/rip-tool/internal/bluray"
	"github.com/teemukurki/rip-tool/internal/common"
	"github.com/teemukurki/rip-tool/internal/utils"
)

var blurayOpts = common.Options{}

func blurayPrecheck() []error {
	var errors []error
	errors = append(errors, utils.CheckCommandAvailable("bd_list_titles", "Run 'sudo apt install libbluray-bin'"))
	errors = append(errors, utils.CheckCommandAvailable("bd_splice", "Run 'sudo apt install libbluray-bin'"))
	errors = append(errors, utils.CheckCommandAvailable("ffmpeg", "Run 'sudo apt install ffmpeg'"))
	return utils.RemoveNil(errors)
}

var blurayCmd = &cobra.Command{
	Use:   "bluray <title>",
	Short: "Rip Bluray",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		precheckErrors := blurayPrecheck()
		if len(precheckErrors) > 0 {
			for _, err := range precheckErrors {
				fmt.Println(err)
			}
			return nil
		} else {
			return bluray.RunBluray(blurayOpts, title)
		}
	},
}

func home() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func init() {
	rootCmd.AddCommand(blurayCmd)

	f := blurayCmd.Flags()
	f.BoolVar(&blurayOpts.Show, "show", false, "TV show mode")
	f.StringVar(&blurayOpts.DiskPath, "disk-path", "/dev/sr0", "Disk path")
	f.IntVar(&blurayOpts.Season, "season", 1, "Season number")
	f.IntVar(&blurayOpts.Disk, "disk", 1, "Disk number")
	f.IntVar(&blurayOpts.MinLength, "min-length", 20, "Min track length (minutes)")
	f.IntVar(&blurayOpts.MaxLength, "max-length", 0, "Max track length (0 = disabled)")
	f.StringVar(&blurayOpts.AudioCodec, "audio-codec", "aac", "Audio codec")
	// Allow passing ffmpeg params as indivitual values for easier use
	f.StringVar(&blurayOpts.VideoEncodingParams, "video-encoding-params", "-c:v h264_nvenc -preset p7 -rc vbr -cq 28", "FFmpeg video params")
	f.IntSliceVar(&blurayOpts.AudioTrack, "audio-track", nil, "Select audio tracks for output")
	f.StringVar(&blurayOpts.AudioLang, "default-audio-lang", "", "Set default audio track by language")

	f.IntSliceVar(&blurayOpts.VideoTrack, "video-track", nil, "Select video tracks for output")
	f.StringVar(&blurayOpts.VideoLang, "default-video-lang", "", "Set default video track by language")
	f.IntSliceVar(&blurayOpts.SubtitleTrack, "subtitle-track", nil, "Select subtitles track for output")
	f.StringVar(&blurayOpts.SubtitleLang, "default-subtitle-lang", "", "Set default subtitle track by language")

	f.IntSliceVarP(&blurayOpts.Titles, "title", "t", nil, "Specific title(s) to rip")
	f.BoolVar(&blurayOpts.NoAutoLength, "no-auto-lenght", false, "Disable automatic track length detection")

	f.StringVar(&blurayOpts.KeyPath, "key", filepath.Join(home(), ".config", "aacs", "KEYDB.cfg"), "Location of aacs KEYDB.cfg")
}
