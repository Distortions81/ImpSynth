package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Distortions81/impsynth"
)

const (
	sampleRate      = 44100
	renderChunkSize = 1024
	releaseTailMS   = 250
)

type noteEvent struct {
	StartMS    int
	DurationMS int
	MIDINote   int
}

type patchFile struct {
	Name string         `json:"name"`
	Regs map[string]int `json:"regs"`
}

func main() {
	inputPath := "examples/twinkle.csv"
	outputPath := "twinkle.wav"
	patchPath := "examples/patches/xylophone.json"
	if len(os.Args) > 1 {
		inputPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}
	if len(os.Args) > 3 {
		patchPath = os.Args[3]
	}

	events, err := loadEvents(inputPath)
	if err != nil {
		fail(err)
	}
	patch, err := loadPatch(patchPath)
	if err != nil {
		fail(err)
	}

	opl := impsynth.New(sampleRate)
	configureVoice(opl, patch)

	pcm, err := renderSong(opl, events)
	if err != nil {
		fail(err)
	}
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fail(err)
		}
	}
	if err := writeStereoS16WAV(outputPath, sampleRate, pcm.Bytes()); err != nil {
		fail(err)
	}

	fmt.Printf("wrote %s from %s using %s (%d notes)\n", outputPath, inputPath, patch.Name, len(events))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func loadEvents(path string) ([]noteEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	filtered, err := filterComments(f)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(filtered))
	r.FieldsPerRecord = 3

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	events := make([]noteEvent, 0, len(records))
	for i, rec := range records {
		startMS, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid start_ms: %w", i+1, err)
		}
		durationMS, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid duration_ms: %w", i+1, err)
		}
		midiNote, err := strconv.Atoi(strings.TrimSpace(rec[2]))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid midi_note: %w", i+1, err)
		}
		if startMS < 0 || durationMS <= 0 || midiNote < 0 || midiNote > 127 {
			return nil, fmt.Errorf("row %d: out-of-range values", i+1)
		}
		events = append(events, noteEvent{
			StartMS:    startMS,
			DurationMS: durationMS,
			MIDINote:   midiNote,
		})
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartMS == events[j].StartMS {
			return events[i].MIDINote < events[j].MIDINote
		}
		return events[i].StartMS < events[j].StartMS
	})
	return events, nil
}

func loadPatch(path string) (*patchFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var patch patchFile
	if err := json.Unmarshal(data, &patch); err != nil {
		return nil, err
	}
	if patch.Name == "" {
		patch.Name = filepath.Base(path)
	}
	if len(patch.Regs) == 0 {
		return nil, fmt.Errorf("patch %s has no registers", path)
	}
	return &patch, nil
}

func filterComments(r io.Reader) (string, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(src), "\n")
	var filtered []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n"), nil
}

func configureVoice(opl *impsynth.Synth, patch *patchFile) {
	opl.Reset()
	opl.WriteReg(0x01, 0x20)
	opl.WriteReg(0xC0, 0x30)

	regNames := make([]string, 0, len(patch.Regs))
	for reg := range patch.Regs {
		regNames = append(regNames, reg)
	}
	sort.Strings(regNames)
	for _, reg := range regNames {
		addr, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(reg), "0x"), 16, 16)
		if err != nil {
			fail(fmt.Errorf("patch %q has invalid register %q", patch.Name, reg))
		}
		value := patch.Regs[reg]
		if value < 0 || value > 0xFF {
			fail(fmt.Errorf("patch %q register %s out of range", patch.Name, reg))
		}
		opl.WriteReg(uint16(addr), uint8(value))
	}
}

func renderSong(opl *impsynth.Synth, events []noteEvent) (*bytes.Buffer, error) {
	var pcm bytes.Buffer
	cursorMS := 0
	for _, event := range events {
		if event.StartMS < cursorMS {
			return nil, fmt.Errorf("overlapping event at %dms; the example player is monophonic", event.StartMS)
		}
		if err := renderSegment(opl, &pcm, event.StartMS-cursorMS); err != nil {
			return nil, err
		}
		keyOn(opl, event.MIDINote)
		if err := renderSegment(opl, &pcm, event.DurationMS); err != nil {
			return nil, err
		}
		keyOff(opl, event.MIDINote)
		cursorMS = event.StartMS + event.DurationMS
	}
	if err := renderSegment(opl, &pcm, releaseTailMS); err != nil {
		return nil, err
	}
	return &pcm, nil
}

func renderSegment(opl *impsynth.Synth, dst io.Writer, durationMS int) error {
	framesRemaining := int(math.Round(float64(durationMS) * sampleRate / 1000.0))
	for framesRemaining > 0 {
		frames := framesRemaining
		if frames > renderChunkSize {
			frames = renderChunkSize
		}
		if err := binary.Write(dst, binary.LittleEndian, opl.GenerateStereoS16(frames)); err != nil {
			return err
		}
		framesRemaining -= frames
	}
	return nil
}

func keyOn(opl *impsynth.Synth, midiNote int) {
	a0, b0 := oplNoteRegs(midiNote)
	opl.WriteReg(0xA0, a0)
	opl.WriteReg(0xB0, b0|0x20)
}

func keyOff(opl *impsynth.Synth, midiNote int) {
	a0, b0 := oplNoteRegs(midiNote)
	opl.WriteReg(0xA0, a0)
	opl.WriteReg(0xB0, b0)
}

func oplNoteRegs(midiNote int) (uint8, uint8) {
	block, fnum := midiToOPL(midiNote)
	return uint8(fnum & 0xFF), uint8(block<<2) | uint8((fnum>>8)&0x03)
}

func midiToOPL(midiNote int) (uint8, uint16) {
	freqHz := 440.0 * math.Pow(2, float64(midiNote-69)/12.0)

	bestBlock := uint8(0)
	bestFNum := uint16(0)
	bestErr := math.MaxFloat64

	for block := uint8(0); block <= 7; block++ {
		scale := math.Pow(2, float64(int(block)-1))
		fnum := int(math.Round(freqHz * 524288.0 / (49716.0 * scale)))
		if fnum < 0 || fnum > 1023 {
			continue
		}
		actualHz := float64(fnum) * scale * 49716.0 / 524288.0
		err := math.Abs(actualHz - freqHz)
		if err < bestErr {
			bestErr = err
			bestBlock = block
			bestFNum = uint16(fnum)
		}
	}

	return bestBlock, bestFNum
}

func writeStereoS16WAV(path string, sampleRate int, pcm []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	const (
		channelCount  = 2
		bitsPerSample = 16
	)
	blockAlign := channelCount * (bitsPerSample / 8)
	byteRate := sampleRate * blockAlign
	riffSize := 36 + len(pcm)

	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(riffSize)); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVEfmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(channelCount)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(byteRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(blockAlign)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return err
	}
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(len(pcm))); err != nil {
		return err
	}
	_, err = f.Write(pcm)
	return err
}
