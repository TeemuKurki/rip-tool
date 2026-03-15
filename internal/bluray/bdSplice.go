package bluray

import (
	"fmt"
	"os/exec"

	"github.com/teemukurki/rip-tool/internal/common"
)

func chapterRange(chapters int) string {
	if chapters == 1 {
		return "1"
	}
	return fmt.Sprintf("1-%d", chapters)
}

func BdSpliceCmd(opts common.Options, track BDTitle) *exec.Cmd {
	args := []string{
		"-t", fmt.Sprintf("%d", track.Index),
		"-c", chapterRange(track.Chapters),
		"-k", opts.KeyPath,
		opts.DiskPath,
	}
	cmd := exec.Command(
		"bd_splice",
		args...,
	)
	fmt.Println(args)
	return cmd
}
