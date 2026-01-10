package utils

import (
	"bufio"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type StreamInfo struct {
	StreamIndex int
	HexID       string
	Type        string // audio | subtitle
}

type StreamMap struct {
	OriginalIndex int
	TargetIndex   int
}

var streamRe = regexp.MustCompile(
	`^\s*Stream #(\d+:\d+)\[(0x[0-9A-Fa-f]+)\]: (Audio|Subtitle):`,
)

var streamMapRe = regexp.MustCompile(
	`^\s*Stream #(\d+:\d+) -> #(\d+:\d+)`,
)

// Extracts Subtitle and Audio stream information from ffmpeg logs
func ParseStreams(r io.Reader, out io.Writer) ([]StreamInfo, error) {
	var streams []StreamInfo
	var mappings []StreamMap

	reader := bufio.NewReader(r)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return streams, err
		}

		// Extact new stream ID after mapping. E.g. "Stream #0:2[0x21]: Subtitle: dvd_subtitle"
		streamInfoLine := streamRe.FindStringSubmatch(line)
		if streamInfoLine == nil {
			// Extract Stream mapping info. E.g. "Stream #0:2 -> #0:1 (copy)"
			mapLine := streamMapRe.FindStringSubmatch(line)
			if mapLine != nil {
				origParts := strings.Split(mapLine[1], ":")   // 0:2
				targetParts := strings.Split(mapLine[2], ":") // 0:1

				origId, _ := strconv.Atoi(origParts[1])     // 2
				targetId, _ := strconv.Atoi(targetParts[1]) // 1

				mappings = append(mappings, StreamMap{
					OriginalIndex: origId,
					TargetIndex:   targetId + 1, // Add +1 to target to make it ready for mkvpropedit
				})
			}
			continue
		}

		parts := strings.Split(streamInfoLine[1], ":") // 0,2
		id, _ := strconv.Atoi(parts[1])                // 2

		streams = append(streams, StreamInfo{
			StreamIndex: id,
			HexID:       streamInfoLine[2],                  //0x21
			Type:        strings.ToLower(streamInfoLine[3]), // Subtitle
		})

	}

	var resolvedStreams []StreamInfo

	for _, streamMap := range mappings {
		originalIndex := slices.IndexFunc(streams, func(s StreamInfo) bool {
			return s.StreamIndex == streamMap.OriginalIndex
		})
		targetIndex := slices.IndexFunc(streams, func(s StreamInfo) bool {
			return s.StreamIndex == streamMap.TargetIndex
		})
		if originalIndex > -1 && targetIndex > -1 {
			// Only resolve streams that have mapping
			resolvedStreams = append(resolvedStreams, StreamInfo{
				Type:        streams[originalIndex].Type,      // Grab original type
				StreamIndex: streams[targetIndex].StreamIndex, // Grab mapping target Stream Index
				HexID:       streams[originalIndex].HexID,     // Grab original HexID. Will resolve to language data later on
			})
		}

	}

	return resolvedStreams, nil
}
