package bluray

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/teemukurki/rip-tool/internal/common"
)

type BDTitle struct {
	Index     int
	Duration  int
	Chapters  int
	Angles    int
	Clips     int
	Playlist  string
	Video     int
	Audio     int
	PG        int
	IG        int
	SV        int
	SA        int
	AudioLang []string
	PGLang    []string
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Return track duration in seconds
func calculateDuration(durationStr string) int {
	splitted := strings.Split(durationStr, ":")
	if len(splitted) != 3 {
		fmt.Println("Invalid duration format. Expected 3 parts got %d. Original value %s", len(splitted), splitted)
		return 0
	}
	fullDuration := 0
	hours := toInt(splitted[0])
	minutes := toInt(splitted[1])
	seconds := toInt(splitted[2])
	if seconds > 0 {
		fullDuration += seconds
	}
	if minutes > 0 {
		fullDuration += minutes * 60
	}
	if hours > 0 {
		fullDuration += hours * 60 * 60
	}

	return fullDuration
}

func bdListTitlesCmd(opts common.Options) *exec.Cmd {
	args := []string{
		"-l",
		opts.DiskPath,
	}
	return exec.Command("bd_list_titles", args...)
}

func GetDBTitles(opts common.Options) ([]BDTitle, error) {
	cmd := bdListTitlesCmd(opts)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run bd_list_titles: %w", err)
	}

	// Extract index lines data
	indexLine := regexp.MustCompile(
		`^index:\s+(\d+)\s+duration:\s+([\d:]+)\s+chapters:\s+(\d+)\s+angles:\s+(\d+)\s+clips:\s+(\d+).*playlist:\s+(\d+\.mpls).*V:(\d+)\s+A:(\d+)\s+PG:(\d+)\s+IG:(\d+)\s+SV:(\d+)\s+SA:(\d+)`,
	)
	// Extract AUD (Audio languages) lines data
	audLine := regexp.MustCompile(`^\s+AUD:\s+(.+)$`)
	// Extract PG (Subtitle languages) lines data
	pgLine := regexp.MustCompile(`^\s+PG\s+:\s+(.+)$`)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	var titles []BDTitle
	var current *BDTitle

	for scanner.Scan() {
		line := scanner.Text()
		// Match "Index" line
		if match := indexLine.FindStringSubmatch(line); match != nil {
			// Add previous indexes data to list
			if current != nil {
				titles = append(titles, *current)
			}
			// Create pointer to new DBTitle struct
			current = &BDTitle{
				Index:    toInt(match[1]),
				Duration: calculateDuration(match[2]),
				Chapters: toInt(match[3]),
				Angles:   toInt(match[4]),
				Clips:    toInt(match[5]),
				Playlist: match[6],
				Video:    toInt(match[7]),
				Audio:    toInt(match[8]),
				PG:       toInt(match[9]),
				IG:       toInt(match[10]),
				SV:       toInt(match[11]),
				SA:       toInt(match[12]),
			}
			continue
		}

		// If there in no DBTitle reference, skip
		if current == nil {
			continue
		}
		// Match "AUD" line
		if m := audLine.FindStringSubmatch(line); m != nil {
			// If found match to audio languages; update current (latest DBTitle struct reference)
			current.AudioLang = strings.Fields(m[1])
		}
		// Match "PG" line
		if m := pgLine.FindStringSubmatch(line); m != nil {
			// If found match to subtitle languages; update current (latest DBTitle struct reference)
			current.PGLang = strings.Fields(m[1])
		}
	}
	// Add last index data to list
	if current != nil {
		titles = append(titles, *current)
	}
	return titles, nil
}
