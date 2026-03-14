package impsynth

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var benchmarkVoiceChannels = []int{0, 1, 2}
var benchmarkMaxVoiceChannels = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}

const (
	exampleSongSampleRate = 44100
	exampleSongChunkSize  = 1024
	exampleSongReleaseMS  = 250
)

type benchmarkNoteEvent struct {
	StartMS    int
	DurationMS int
	MIDINote   int
}

type benchmarkPatchFile struct {
	Name string         `json:"name"`
	Regs map[string]int `json:"regs"`
}

type synthReferenceCorpusCase struct {
	name string
	regs []uint16
}

var synthReferenceCorpusCases = []synthReferenceCorpusCase{
	{
		name: "melodic_fm",
		regs: []uint16{0x20, 0x21, 0x23, 0x01, 0x40, 0x08, 0x43, 0x00, 0x60, 0xF2, 0x63, 0xF2, 0x80, 0x24, 0x83, 0x24, 0xC0, 0x30, 0xA0, 0x98, 0xB0, 0x31},
	},
	{
		name: "bright_feedback",
		regs: []uint16{0x20, 0x21, 0x23, 0x21, 0x40, 0x04, 0x43, 0x00, 0x60, 0xF4, 0x63, 0xF4, 0x80, 0x22, 0x83, 0x22, 0xC0, 0x3C, 0xA0, 0xC0, 0xB0, 0x35},
	},
	{
		name: "trem_vib",
		regs: []uint16{0xBD, 0xC0, 0x20, 0xC1, 0x23, 0xC1, 0x40, 0x18, 0x43, 0x00, 0x60, 0xF3, 0x63, 0xF3, 0x80, 0x34, 0x83, 0x34, 0xC0, 0x30, 0xA0, 0x88, 0xB0, 0x33},
	},
}

func writeBenchmarkVoice(opl *Synth, ch int) {
	base := uint16(0)
	localCh := ch
	if ch >= 9 {
		base = 0x100
		localCh = ch - 9
	}

	modSlots := [9]uint16{0, 1, 2, 8, 9, 10, 16, 17, 18}
	carSlots := [9]uint16{3, 4, 5, 11, 12, 13, 19, 20, 21}
	mod := modSlots[localCh]
	car := carSlots[localCh]

	opl.WriteReg(base+0x20+mod, 0x01)
	opl.WriteReg(base+0x20+car, 0x01)
	opl.WriteReg(base+0x40+mod, 0x18)
	opl.WriteReg(base+0x40+car, 0x00)
	opl.WriteReg(base+0x60+mod, 0xF4)
	opl.WriteReg(base+0x60+car, 0xF6)
	opl.WriteReg(base+0x80+mod, 0x55)
	opl.WriteReg(base+0x80+car, 0x14)
	opl.WriteReg(base+0xC0+uint16(localCh), 0x30)
	opl.WriteReg(base+0xA0+uint16(localCh), 0x98)
	opl.WriteReg(base+0xB0+uint16(localCh), 0x31)
}

func benchmarkSynth(sampleRate int, channels []int) *Synth {
	opl := New(sampleRate)
	opl.WriteReg(0x01, 0x20)
	for _, ch := range channels {
		writeBenchmarkVoice(opl, ch)
	}
	return opl
}

func benchmarkReferenceCorpusSynth(sampleRate int, regs []uint16) *Synth {
	opl := New(sampleRate)
	opl.WriteReg(0x01, 0x20)
	for i := 0; i+1 < len(regs); i += 2 {
		opl.WriteReg(regs[i], uint8(regs[i+1]))
	}
	return opl
}

func mustLoadExampleSongBenchmarkData(b *testing.B) ([]benchmarkNoteEvent, *benchmarkPatchFile) {
	b.Helper()

	eventsData, err := os.ReadFile(filepath.Join("examples", "twinkle.csv"))
	if err != nil {
		b.Fatalf("read example song: %v", err)
	}
	patchData, err := os.ReadFile(filepath.Join("examples", "patches", "xylophone.json"))
	if err != nil {
		b.Fatalf("read example patch: %v", err)
	}

	events, err := parseExampleSongEvents(eventsData)
	if err != nil {
		b.Fatalf("parse example song: %v", err)
	}
	patch, err := parseExampleSongPatch(patchData)
	if err != nil {
		b.Fatalf("parse example patch: %v", err)
	}
	return events, patch
}

func parseExampleSongEvents(src []byte) ([]benchmarkNoteEvent, error) {
	lines := strings.Split(string(src), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		filtered = append(filtered, line)
	}

	r := csv.NewReader(strings.NewReader(strings.Join(filtered, "\n")))
	r.FieldsPerRecord = 3
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	events := make([]benchmarkNoteEvent, 0, len(records))
	for _, rec := range records {
		startMS, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			return nil, err
		}
		durationMS, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			return nil, err
		}
		midiNote, err := strconv.Atoi(strings.TrimSpace(rec[2]))
		if err != nil {
			return nil, err
		}
		if startMS < 0 || durationMS <= 0 || midiNote < 0 || midiNote > 127 {
			return nil, strconv.ErrSyntax
		}
		events = append(events, benchmarkNoteEvent{
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

func parseExampleSongPatch(src []byte) (*benchmarkPatchFile, error) {
	var patch benchmarkPatchFile
	if err := json.Unmarshal(src, &patch); err != nil {
		return nil, err
	}
	if patch.Name == "" {
		patch.Name = "xylophone"
	}
	return &patch, nil
}

func configureExampleSongVoice(opl *Synth, patch *benchmarkPatchFile) {
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
			panic(err)
		}
		opl.WriteReg(uint16(addr), uint8(patch.Regs[reg]))
	}
}

func benchmarkRenderExampleSong(opl *Synth, events []benchmarkNoteEvent) {
	cursorMS := 0
	for _, event := range events {
		renderExampleSongSegment(opl, event.StartMS-cursorMS)
		keyOnExampleSong(opl, event.MIDINote)
		renderExampleSongSegment(opl, event.DurationMS)
		keyOffExampleSong(opl, event.MIDINote)
		cursorMS = event.StartMS + event.DurationMS
	}
	renderExampleSongSegment(opl, exampleSongReleaseMS)
}

func benchmarkRenderExampleSongFastSilenceFill(opl *Synth, events []benchmarkNoteEvent) {
	cursorMS := 0
	for _, event := range events {
		renderExampleSongSegment(opl, event.StartMS-cursorMS)
		keyOnExampleSong(opl, event.MIDINote)
		renderExampleSongSegment(opl, event.DurationMS)
		keyOffExampleSong(opl, event.MIDINote)
		cursorMS = event.StartMS + event.DurationMS
	}
	renderExampleSongReleaseTailFastSilence(opl, exampleSongReleaseMS)
}

func renderExampleSongSegment(opl *Synth, durationMS int) {
	framesRemaining := int(math.Round(float64(durationMS) * exampleSongSampleRate / 1000.0))
	for framesRemaining > 0 {
		frames := framesRemaining
		if frames > exampleSongChunkSize {
			frames = exampleSongChunkSize
		}
		_ = opl.GenerateStereoS16(frames)
		framesRemaining -= frames
	}
}

func renderExampleSongReleaseTailFastSilence(opl *Synth, durationMS int) {
	framesRemaining := int(math.Round(float64(durationMS) * exampleSongSampleRate / 1000.0))
	for framesRemaining > 0 {
		frames := framesRemaining
		if frames > exampleSongChunkSize {
			frames = exampleSongChunkSize
		}
		if benchmarkSynthOutputSilent(opl) {
			_ = benchmarkFillStereoSilence(opl, frames)
		} else {
			_ = opl.GenerateStereoS16(frames)
		}
		framesRemaining -= frames
	}
}

func benchmarkSynthOutputSilent(opl *Synth) bool {
	return opl.activeMask == 0 &&
		(!opl.resamplePrimed ||
			(opl.resamplePrevL == 0 &&
				opl.resamplePrevR == 0 &&
				opl.resampleNextL == 0 &&
				opl.resampleNextR == 0))
}

func benchmarkFillStereoSilence(opl *Synth, frames int) []int16 {
	if frames <= 0 {
		return nil
	}
	need := frames * 2
	if cap(opl.stereoBuf) < need {
		opl.stereoBuf = make([]int16, need)
	} else {
		opl.stereoBuf = opl.stereoBuf[:need]
	}
	clear(opl.stereoBuf)
	return opl.stereoBuf
}

func keyOnExampleSong(opl *Synth, midiNote int) {
	a0, b0 := oplNoteRegsForBenchmark(midiNote)
	opl.WriteReg(0xA0, a0)
	opl.WriteReg(0xB0, b0|0x20)
}

func keyOffExampleSong(opl *Synth, midiNote int) {
	a0, b0 := oplNoteRegsForBenchmark(midiNote)
	opl.WriteReg(0xA0, a0)
	opl.WriteReg(0xB0, b0)
}

func oplNoteRegsForBenchmark(midiNote int) (uint8, uint8) {
	block, fnum := midiToOPLForBenchmark(midiNote)
	return uint8(fnum & 0xFF), uint8(block<<2) | uint8((fnum>>8)&0x03)
}

func midiToOPLForBenchmark(midiNote int) (uint8, uint16) {
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
		actualHz := float64(fnum) * 49716.0 * scale / 524288.0
		err := math.Abs(actualHz - freqHz)
		if err < bestErr {
			bestErr = err
			bestBlock = block
			bestFNum = uint16(fnum)
		}
	}
	return bestBlock, bestFNum
}

func TestGenerateStereoS16ProducesPCM(t *testing.T) {
	opl := New(49716)
	opl.WriteReg(0x01, 0x20)
	opl.WriteReg(0x20, 0x01)
	opl.WriteReg(0x23, 0x01)
	opl.WriteReg(0x60, 0xF3)
	opl.WriteReg(0x63, 0xF3)
	opl.WriteReg(0x80, 0x24)
	opl.WriteReg(0x83, 0x24)
	opl.WriteReg(0xA0, 0x98)
	opl.WriteReg(0xB0, 0x31)
	opl.WriteReg(0x43, 0x00)
	opl.WriteReg(0xC0, 0x30)
	pcm := opl.GenerateStereoS16(256)
	if len(pcm) != 512 {
		t.Fatalf("samples=%d want=512", len(pcm))
	}
	nonZero := false
	for _, s := range pcm {
		if s != 0 {
			nonZero = true
			break
		}
	}
	if !nonZero {
		t.Fatal("expected audible PCM")
	}
}

func TestGenerateStereoS16ReusesBuffer(t *testing.T) {
	opl := New(49716)
	opl.WriteReg(0x20, 0x01)
	opl.WriteReg(0x23, 0x01)
	opl.WriteReg(0xA0, 0x98)
	opl.WriteReg(0xB0, 0x31)
	opl.WriteReg(0x43, 0x00)
	_ = opl.GenerateStereoS16(256)
	allocs := testing.AllocsPerRun(100, func() {
		_ = opl.GenerateStereoS16(256)
	})
	if allocs != 0 {
		t.Fatalf("GenerateStereoS16 allocs=%v want 0", allocs)
	}
}

func BenchmarkGenerateStereoS16_2048Frames(b *testing.B) {
	opl := benchmarkSynth(49716, benchmarkVoiceChannels)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}

func BenchmarkGenerateStereoS16_2048Frames_44100Hz(b *testing.B) {
	opl := benchmarkSynth(44100, benchmarkVoiceChannels)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}

func BenchmarkGenerateStereoS16_2048Frames_MaxVoices(b *testing.B) {
	opl := benchmarkSynth(49716, benchmarkMaxVoiceChannels)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}

func BenchmarkGenerateStereoS16_2048Frames_MaxVoices_44100Hz(b *testing.B) {
	opl := benchmarkSynth(44100, benchmarkMaxVoiceChannels)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}

func BenchmarkRenderExampleSongTwinkle44100Hz(b *testing.B) {
	events, patch := mustLoadExampleSongBenchmarkData(b)
	totalDurationMS := exampleSongReleaseMS
	if len(events) > 0 {
		last := events[len(events)-1]
		totalDurationMS += last.StartMS + last.DurationMS
	}
	totalFrames := int(math.Round(float64(totalDurationMS) * exampleSongSampleRate / 1000.0))

	b.ReportAllocs()
	b.SetBytes(int64(totalFrames * 2 * 2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opl := New(exampleSongSampleRate)
		configureExampleSongVoice(opl, patch)
		benchmarkRenderExampleSong(opl, events)
	}
}

func BenchmarkRenderExampleSongTwinkle44100Hz_FastSilenceFill(b *testing.B) {
	events, patch := mustLoadExampleSongBenchmarkData(b)
	totalDurationMS := exampleSongReleaseMS
	if len(events) > 0 {
		last := events[len(events)-1]
		totalDurationMS += last.StartMS + last.DurationMS
	}
	totalFrames := int(math.Round(float64(totalDurationMS) * exampleSongSampleRate / 1000.0))

	b.ReportAllocs()
	b.SetBytes(int64(totalFrames * 2 * 2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opl := New(exampleSongSampleRate)
		configureExampleSongVoice(opl, patch)
		benchmarkRenderExampleSongFastSilenceFill(opl, events)
	}
}

func BenchmarkGenerateStereoS16_2048Frames_ReferenceCorpus(b *testing.B) {
	for _, tc := range synthReferenceCorpusCases {
		b.Run(tc.name, func(b *testing.B) {
			opl := benchmarkReferenceCorpusSynth(49716, tc.regs)
			b.ReportAllocs()
			b.SetBytes(2048 * 2 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = opl.GenerateStereoS16(2048)
			}
		})
	}
}

func BenchmarkGenerateStereoS16_2048Frames_ReferenceCorpus_44100Hz(b *testing.B) {
	for _, tc := range synthReferenceCorpusCases {
		b.Run(tc.name, func(b *testing.B) {
			opl := benchmarkReferenceCorpusSynth(44100, tc.regs)
			b.ReportAllocs()
			b.SetBytes(2048 * 2 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = opl.GenerateStereoS16(2048)
			}
		})
	}
}
