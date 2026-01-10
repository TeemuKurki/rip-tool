package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teemukurki/rip-tool/internal/bluray"
	"github.com/teemukurki/rip-tool/internal/common"
)

var blurayOpts = common.DefaultOptions()

// blurayCmd represents the bluray command
var blurayCmd = &cobra.Command{
	Use:   "bluray <title>",
	Short: "Rip DVD",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		precheckErrors := dvdPrecheck()
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

func init() {
	rootCmd.AddCommand(blurayCmd)

	f := blurayCmd.Flags()
	f.BoolVar(&blurayOpts.Show, "show", false, "TV show mode")
	f.StringVar(&blurayOpts.DiskPath, "disk-path", blurayOpts.DiskPath, "Disk path")
	f.IntVar(&blurayOpts.Season, "season", blurayOpts.Season, "Season number")
	f.IntVar(&blurayOpts.Disk, "disk", blurayOpts.Disk, "Disk number")
	f.IntVar(&blurayOpts.MinLength, "min-length", blurayOpts.MinLength, "Min track length (minutes)")
	f.IntVar(&blurayOpts.MaxLength, "max-length", blurayOpts.MaxLength, "Max track length (0 = disabled)")
	f.StringVar(&blurayOpts.AudioCodec, "audio-codec", blurayOpts.AudioCodec, "Audio codec")
	f.StringVar(&blurayOpts.VideoEncodingParams, "video-encoding-params", blurayOpts.VideoEncodingParams, "FFmpeg video params")
	f.IntVar(&blurayOpts.AudioTrack, "audio-track", blurayOpts.AudioTrack, "Select single audio track for output")
	f.StringVar(&blurayOpts.AudioLang, "default-audio-lang", blurayOpts.AudioLang, "Set default audio track by language")

	f.IntVar(&blurayOpts.VideoTrack, "video-track", blurayOpts.VideoTrack, "Select single video track for output")
	f.StringVar(&blurayOpts.VideoLang, "default-video-lang", blurayOpts.VideoLang, "Set default video track by language")
	f.IntVar(&blurayOpts.SubtitleTrack, "subtitle-track", blurayOpts.SubtitleTrack, "Select single subtitle track for output")
	f.StringVar(&blurayOpts.SubtitleLang, "default-subtitle-lang", blurayOpts.SubtitleLang, "Set default subtitle track by language")

	f.IntSliceVarP(&blurayOpts.Titles, "title", "t", nil, "Specific title(s) to rip")
	f.BoolVar(&blurayOpts.NoAutoLength, "no-auto-lenght", blurayOpts.NoAutoLength, "Disable automatic track length detection")

	f.StringVar(&blurayOpts.KeyPath, "--key", blurayOpts.KeyPath, "Location of aacs KEYDB.cfg")
}
