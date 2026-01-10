package dvd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
	"unicode/utf8"

	"github.com/teemukurki/rip-tool/internal/common"
)

type LSDVD struct {
	XMLName      xml.Name `xml:"lsdvd"`
	Device       string   `xml:"device"`
	Title        string   `xml:"title"`
	VMGID        string   `xml:"vmg_id"`
	ProviderID   string   `xml:"provider_id"`
	Tracks       []Track  `xml:"track"`
	LongestTrack int      `xml:"longest_track"`
}

type Track struct {
	Index   int       `xml:"ix"`
	Length  float32   `xml:"length"`
	VTSID   string    `xml:"vts_id"`
	Subps   []Subp    `xml:"subp"`
	Chapter []Chapter `xml:"chapter"`
	Audio   []Audio   `xml:"audio"`
}

type Subp struct {
	Index    int    `xml:"ix"`
	LangCode string `xml:"langcode"`
	Language string `xml:"language"`
	Content  string `xml:"content"`
	StreamID string `xml:"streamid"`
}

type Chapter struct {
	Index     int     `xml:"ix"`
	Length    float32 `xml:"length"`
	Startcell int     `xml:"startcell"`
}

type Audio struct {
	Index        int    `xml:"ix"`
	LangCode     string `xml:"langcode"`
	Language     string `xml:"language"`
	Format       string `xml:"format"`
	Frequency    int    `xml:"frequency"`
	Quantization string `xml:"quantization"`
	Channels     int    `xml:"channels"`
	APMode       int    `xml:"ap_mode"`
	Content      string `xml:"content"`
	StreamID     string `xml:"streamid"`
}

func lsdvdCmd(opts common.Options) *exec.Cmd {
	args := []string{
		"-acs",
		opts.DiskPath,
		"-Ox",
	}
	return exec.Command("lsdvd", args...)
}

// Table of valid characters in xml 1.0. See https://www.baeldung.com/java-xml-invalid-characters
func isValidXMLChar(r rune) bool {
	return r == 0x9 || r == 0xA || r == 0xD || // TAB | LF | CR
		(r >= 0x20 && r <= 0xD7FF) || // All alpha-numeric
		(r >= 0xE000 && r <= 0xFFFD) || // SMP (Supplementary Multilingual Plane)
		(r >= 0x10000 && r <= 0x10FFFF) // BMP (Basic Multilingual Plane)
}

func stripIllegalXML(data []byte) []byte {
	xmlStartIndex := bytes.Index(data, []byte("<?xml"))
	if xmlStartIndex > 0 {
		fmt.Println("lsdvd output contains error. Ignoring errors", string(data[:xmlStartIndex]))
	}
	xmlData := data[xmlStartIndex:]

	var buf bytes.Buffer
	for len(xmlData) > 0 {
		r, size := utf8.DecodeRune(xmlData)
		if r == utf8.RuneError && size == 1 {
			// skip invalid UTF-8 byte
			xmlData = xmlData[size:]
			continue
		}
		if isValidXMLChar(r) {
			// If cahacter found in XML table. Add character to buffer
			buf.WriteRune(r)
		}
		xmlData = xmlData[size:]
	}
	return buf.Bytes()
}

func GetLsdvdInfo(opts common.Options) (*LSDVD, error) {
	cmd := lsdvdCmd(opts)

	// Run the command and capture output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run lsdvd: %w", err)
	}

	// Strip non-xml compliant characters
	illigalCharsStrip := stripIllegalXML(output)

	// Unmarshal XML into LSDVD struct
	var lsdvd LSDVD
	if err := xml.Unmarshal(illigalCharsStrip, &lsdvd); err != nil {
		return nil, fmt.Errorf("failed to parse lsdvd XML %w", err)
	}

	return &lsdvd, nil
}

var EmptyLsdvd = LSDVD{}
