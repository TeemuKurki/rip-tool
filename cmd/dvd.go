package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teemukurki/rip-tool/internal/common"
	"github.com/teemukurki/rip-tool/internal/dvd"
	"github.com/teemukurki/rip-tool/internal/utils"
)

func dvdPrecheck() []error {
	var errors []error
	errors = append(errors, utils.CheckCommandAvailable("ffmpeg", "Run 'sudo apt install ffmpeg'"))
	errors = append(errors, utils.CheckCommandAvailable("lsdvd", "Run 'sudo apt install lsdvd'"))
	errors = append(errors, utils.CheckCommandAvailable("mpv", "Run 'sudo apt install mpv'"))
	errors = append(errors, utils.CheckCommandAvailable("mkvpropedit", "Run 'sudo apt install mkvtoolnix'"))
	return utils.RemoveNil(errors)
}

var dvdOpts = common.Options{}

var dvdCmd = &cobra.Command{
	Use:       "dvd <title>",
	Short:     "Rip DVD",
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"title"},
	RunE: func(cmd *cobra.Command, args []string) error {

		title := ""
		if len(args) > 0 {
			title = args[0]
		}
		precheckErrors := dvdPrecheck()
		if len(precheckErrors) > 0 {
			for _, err := range precheckErrors {
				fmt.Println(err)
			}
			return nil
		} else {
			return dvd.RunDVD(dvdOpts, title)
		}
	},
}

func init() {
	rootCmd.AddCommand(dvdCmd)

	f := dvdCmd.Flags()
	f.BoolVar(&dvdOpts.Show, "show", false, "TV show mode")
	f.StringVar(&dvdOpts.DiskPath, "disk-path", "/dev/sr0", "Disk path")
	f.IntVar(&dvdOpts.Season, "season", 1, "Season number")
	f.IntVar(&dvdOpts.Disk, "disk", 1, "Disk number")
	f.IntVar(&dvdOpts.MinLength, "min-length", 20, "Min track length (minutes)")
	f.IntVar(&dvdOpts.MaxLength, "max-length", 0, "Max track length (0 = disabled)")
	f.StringVar(&dvdOpts.AudioCodec, "audio-codec", "aac", "Audio codec")
	f.StringVar(&dvdOpts.VideoEncodingParams, "video-encoding-params", "-c:v h264_nvenc -preset p7 -rc vbr -cq 28", "FFmpeg video params")
	f.IntVar(&dvdOpts.AudioTrack, "audio-track", -1, "Select single audio track for output")
	f.StringVar(&dvdOpts.AudioLang, "default-audio-lang", "", "Set default audio track by language")

	f.IntVar(&dvdOpts.VideoTrack, "video-track", -1, "Select single video track for output")
	f.StringVar(&dvdOpts.VideoLang, "default-video-lang", "", "Set default video track by language")
	f.IntVar(&dvdOpts.SubtitleTrack, "subtitle-track", -1, "Select single subtitle track for output")
	f.StringVar(&dvdOpts.SubtitleLang, "default-subtitle-lang", "", "Set default subtitle track by language")

	f.IntSliceVarP(&dvdOpts.Titles, "title", "t", nil, "Specific title(s) to rip")
	f.BoolVar(&dvdOpts.NoAutoLength, "no-auto-lenght", false, "Disable automatic track length detection")

}
