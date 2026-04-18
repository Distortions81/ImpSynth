package main

import (
	"archive/zip"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	wl1StartMusicChunk    = 261
	wl1MusicCount         = 27
	musicChunkTrailerSize = 88
)

type musicEvent struct {
	Reg   uint16 `json:"reg"`
	Value uint8  `json:"value"`
	Delay uint16 `json:"delay"`
}

type musicChunk struct {
	Name   string       `json:"name"`
	Events []musicEvent `json:"events"`
}

type manifestEntry struct {
	Song  int    `json:"song"`
	Name  string `json:"name"`
	File  string `json:"file"`
	Count int    `json:"event_count"`
}

type manifest struct {
	Source       string          `json:"source"`
	Format       string          `json:"format"`
	MissingSongs []int           `json:"missing_songs,omitempty"`
	Songs        []manifestEntry `json:"songs"`
}

func main() {
	var zipPath string
	var outDir string

	flag.StringVar(&zipPath, "zip", "", "path to wolf3d-shareware.zip")
	flag.StringVar(&outDir, "out", "", "output directory")
	flag.Parse()

	if zipPath == "" || outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: export-wolf3d-shareware-music -zip <zipfile> -out <dir>")
		os.Exit(2)
	}

	audioHead, err := readZipFile(zipPath, "AUDIOHED.WL1")
	if err != nil {
		fail(err)
	}
	audioT, err := readZipFile(zipPath, "AUDIOT.WL1")
	if err != nil {
		fail(err)
	}
	offsets, err := parseAudioHead(audioHead)
	if err != nil {
		fail(err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail(err)
	}

	m := manifest{
		Source: filepath.Base(zipPath),
		Format: "reg_le16,value_u8,delay_le16",
		Songs:  make([]manifestEntry, 0, wl1MusicCount),
	}

	for song := 0; song < wl1MusicCount; song++ {
		chunkIndex := wl1StartMusicChunk + song
		raw, err := loadAudioChunk(audioT, offsets, chunkIndex)
		if err != nil {
			fail(fmt.Errorf("load song %d: %w", song, err))
		}
		chunk, err := parseMusicChunk(raw)
		if err != nil {
			if isUnavailableMusicChunk(err) {
				m.MissingSongs = append(m.MissingSongs, song)
				continue
			}
			fail(fmt.Errorf("parse song %d: %w", song, err))
		}
		base := fmt.Sprintf("%02d-%s.seq", song, slugify(chunk.Name))
		path := filepath.Join(outDir, base)
		if err := writeSeqFile(path, chunk.Events); err != nil {
			fail(fmt.Errorf("write %s: %w", path, err))
		}
		m.Songs = append(m.Songs, manifestEntry{
			Song:  song,
			Name:  chunk.Name,
			File:  base,
			Count: len(chunk.Events),
		})
	}

	manifestPath := filepath.Join(outDir, "manifest.json")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fail(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func readZipFile(zipPath, name string) ([]byte, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if !strings.EqualFold(filepath.Base(f.Name), name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("%s not found in %s", name, zipPath)
}

func parseAudioHead(data []byte) ([]uint32, error) {
	if len(data) < 8 || len(data)%4 != 0 {
		return nil, fmt.Errorf("AUDIOHED invalid size: %d", len(data))
	}
	offsets := make([]uint32, len(data)/4)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[i*4:])
	}
	return offsets, nil
}

func loadAudioChunk(audioT []byte, offsets []uint32, chunk int) ([]byte, error) {
	if chunk < 0 || chunk >= len(offsets)-1 {
		return nil, fmt.Errorf("audio chunk %d out of range", chunk)
	}
	start := offsets[chunk]
	end := offsets[chunk+1]
	if end < start || end > uint32(len(audioT)) {
		return nil, fmt.Errorf("audio chunk %d out of bounds", chunk)
	}
	buf := make([]byte, int(end-start))
	copy(buf, audioT[start:end])
	return buf, nil
}

func parseMusicChunk(raw []byte) (*musicChunk, error) {
	if len(raw) < musicChunkTrailerSize+6 {
		return nil, fmt.Errorf("music chunk unavailable")
	}
	dataLen := int(binary.LittleEndian.Uint16(raw[:2]))
	if dataLen < 4 {
		return nil, fmt.Errorf("invalid data length %d", dataLen)
	}
	if 2+dataLen+musicChunkTrailerSize != len(raw) {
		return nil, fmt.Errorf("unexpected chunk size %d for data length %d", len(raw), dataLen)
	}
	eventBytes := raw[6 : 2+dataLen]
	if len(eventBytes)%4 != 0 {
		return nil, fmt.Errorf("invalid event byte count %d", len(eventBytes))
	}

	events := make([]musicEvent, 0, len(eventBytes)/4)
	for i := 0; i < len(eventBytes); i += 4 {
		regValue := binary.LittleEndian.Uint16(eventBytes[i : i+2])
		delay := binary.LittleEndian.Uint16(eventBytes[i+2 : i+4])
		events = append(events, musicEvent{
			Reg:   uint16(byte(regValue)),
			Value: uint8(regValue >> 8),
			Delay: delay,
		})
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("music chunk has no events")
	}

	return &musicChunk{
		Name:   parseMusicChunkName(raw[len(raw)-musicChunkTrailerSize:]),
		Events: events,
	}, nil
}

func isUnavailableMusicChunk(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unavailable")
}

func parseMusicChunkName(trailer []byte) string {
	if len(trailer) == 0 {
		return ""
	}
	for i := 0; i < len(trailer)-4; i++ {
		if trailer[i] == 0 {
			continue
		}
		j := i
		for j < len(trailer) && trailer[j] != 0 {
			j++
		}
		if j+4 <= len(trailer) && string(trailer[j:j+4]) == "\x00IMF" {
			return string(trailer[i:j])
		}
	}
	return ""
}

func writeSeqFile(path string, events []musicEvent) error {
	buf := make([]byte, 0, len(events)*5)
	for _, ev := range events {
		buf = binary.LittleEndian.AppendUint16(buf, ev.Reg)
		buf = append(buf, ev.Value)
		buf = binary.LittleEndian.AppendUint16(buf, ev.Delay)
	}
	return os.WriteFile(path, buf, 0o644)
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "untitled"
	}
	name = nonSlug.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return "untitled"
	}
	return name
}
