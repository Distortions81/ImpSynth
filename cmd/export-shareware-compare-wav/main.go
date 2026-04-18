package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	impsynth "github.com/Distortions81/impsynth"
	"github.com/remeh/sizedwaitgroup"
)

const (
	sampleRate   = 49716
	wolfTickRate = 700
	doomTickRate = 140
)

type manifestSong struct {
	Name string `json:"name,omitempty"`
	Lump string `json:"lump,omitempty"`
	File string `json:"file"`
}

type manifest struct {
	Source  string         `json:"source"`
	TicRate int            `json:"tic_rate,omitempty"`
	Songs   []manifestSong `json:"songs"`
}

type corpusConfig struct {
	title           string
	slug            string
	manifestPath    string
	baseDir         string
	defaultTickRate int
	newSynth        func(int) *impsynth.Synth
}

type renderJob struct {
	cfg      corpusConfig
	tickRate int
	song     manifestSong
	outPath  string
}

var (
	nukedToolOnce sync.Once
	nukedToolPath string
	nukedToolErr  error
)

func main() {
	var outDir string
	flag.StringVar(&outDir, "out", filepath.Join("out", "shareware-compare-wav"), "output directory")
	flag.Parse()

	configs := []corpusConfig{
		{
			title:           "Wolf3D Shareware",
			slug:            "wolf3d-shareware",
			manifestPath:    filepath.Join("testdata", "wolf3d-shareware-music", "manifest.json"),
			baseDir:         filepath.Join("testdata", "wolf3d-shareware-music"),
			defaultTickRate: wolfTickRate,
			newSynth:        func(rate int) *impsynth.Synth { return impsynth.NewOPL2(rate) },
		},
		{
			title:           "DOOM Shareware",
			slug:            "doom-shareware",
			manifestPath:    filepath.Join("testdata", "doom-shareware-music", "manifest.json"),
			baseDir:         filepath.Join("testdata", "doom-shareware-music"),
			defaultTickRate: doomTickRate,
			newSynth:        func(rate int) *impsynth.Synth { return impsynth.New(rate) },
		},
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail(err)
	}

	jobs, err := buildJobs(configs, outDir)
	if err != nil {
		fail(err)
	}
	if err := runJobs(jobs); err != nil {
		fail(err)
	}

	fmt.Println(outDir)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func buildJobs(configs []corpusConfig, outDir string) ([]renderJob, error) {
	jobs := make([]renderJob, 0, 32)
	for _, cfg := range configs {
		m, err := loadManifest(cfg.manifestPath)
		if err != nil {
			return nil, err
		}
		tickRate := m.TicRate
		if tickRate <= 0 {
			tickRate = cfg.defaultTickRate
		}
		corpusOutDir := filepath.Join(outDir, cfg.slug)
		if err := os.MkdirAll(corpusOutDir, 0o755); err != nil {
			return nil, err
		}
		for _, song := range m.Songs {
			label := songLabel(song)
			jobs = append(jobs, renderJob{
				cfg:      cfg,
				tickRate: tickRate,
				song:     song,
				outPath:  filepath.Join(corpusOutDir, safeSlug(label)+".wav"),
			})
		}
	}
	return jobs, nil
}

func runJobs(jobs []renderJob) error {
	limit := min(max(1, len(jobs)), max(1, runtimeWorkers()/4))
	if limit < 1 {
		limit = 1
	}
	swg := sizedwaitgroup.New(limit)
	errCh := make(chan error, len(jobs))

	for _, job := range jobs {
		job := job
		swg.Add()
		go func() {
			defer swg.Done()
			if err := runJob(job); err != nil {
				errCh <- fmt.Errorf("%s %s: %w", job.cfg.title, songLabel(job.song), err)
			}
		}()
	}

	swg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func runJob(job renderJob) error {
	seqPath := filepath.Join(job.cfg.baseDir, job.song.File)
	frames, err := seqTotalFrames(seqPath, job.tickRate)
	if err != nil {
		return err
	}
	got, err := renderImpSynthSeq(sampleRate, job.tickRate, frames, seqPath, job.cfg.newSynth)
	if err != nil {
		return err
	}
	want, err := renderNukedSeq(sampleRate, job.tickRate, frames, seqPath)
	if err != nil {
		return err
	}
	if len(got) != len(want) {
		return fmt.Errorf("pcm size mismatch: impsynth=%d nuked=%d", len(got), len(want))
	}
	mixed := make([]int16, len(got))
	for i := 0; i+1 < len(got); i += 2 {
		// Left = Nuked mono, Right = ImpSynth mono.
		nm := int16((int(want[i]) + int(want[i+1])) / 2)
		im := int16((int(got[i]) + int(got[i+1])) / 2)
		mixed[i] = nm
		mixed[i+1] = im
	}
	return writeStereoS16WAV(job.outPath, sampleRate, mixed)
}

func loadManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read manifest %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	return m, nil
}

func songLabel(song manifestSong) string {
	if name := strings.TrimSpace(song.Name); name != "" {
		return name
	}
	if lump := strings.TrimSpace(song.Lump); lump != "" {
		return lump
	}
	return strings.TrimSuffix(filepath.Base(song.File), filepath.Ext(song.File))
}

func safeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "untitled"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "untitled"
	}
	return out
}

func seqTotalFrames(seqPath string, tickRate int) (int, error) {
	events, err := readSeq(seqPath)
	if err != nil {
		return 0, err
	}
	if tickRate <= 0 {
		return 0, fmt.Errorf("invalid tick rate %d", tickRate)
	}
	tickFrames := sampleRate / tickRate
	if tickFrames < 1 {
		tickFrames = 1
	}
	var totalTicks uint64
	for _, ev := range events {
		totalTicks += uint64(ev.Delay)
	}
	totalFrames := int(totalTicks) * tickFrames
	if totalFrames < 1 {
		totalFrames = tickFrames
	}
	return totalFrames, nil
}

type seqEvent struct {
	Reg   uint16
	Value uint8
	Delay uint16
}

func readSeq(path string) ([]seqEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data)%5 != 0 {
		return nil, fmt.Errorf("seq size %d is not a multiple of 5", len(data))
	}
	events := make([]seqEvent, 0, len(data)/5)
	for i := 0; i < len(data); i += 5 {
		events = append(events, seqEvent{
			Reg:   binary.LittleEndian.Uint16(data[i : i+2]),
			Value: data[i+2],
			Delay: binary.LittleEndian.Uint16(data[i+3 : i+5]),
		})
	}
	return events, nil
}

func renderImpSynthSeq(sampleRate, tickRate, frames int, seqPath string, newSynth func(int) *impsynth.Synth) ([]int16, error) {
	events, err := readSeq(seqPath)
	if err != nil {
		return nil, err
	}
	opl := newSynth(sampleRate)
	pcm := make([]int16, 0, frames*2)
	eventIndex := 0
	framesUntilNext := 0
	tickFrames := sampleRate / tickRate
	if tickFrames < 1 {
		tickFrames = 1
	}
	for remaining := frames; remaining > 0; {
		if eventIndex >= len(events) && framesUntilNext <= 0 {
			framesUntilNext = remaining
		}
		consumedImmediate := 0
		for framesUntilNext <= 0 && eventIndex < len(events) {
			ev := events[eventIndex]
			eventIndex++
			opl.WriteReg(ev.Reg, ev.Value)
			framesUntilNext = int(ev.Delay) * tickFrames
			if framesUntilNext > 0 {
				break
			}
			consumedImmediate++
			if consumedImmediate >= len(events) {
				break
			}
		}
		if (consumedImmediate >= len(events) || eventIndex >= len(events)) && framesUntilNext <= 0 {
			framesUntilNext = remaining
		}
		chunk := remaining
		if framesUntilNext > 0 && chunk > framesUntilNext {
			chunk = framesUntilNext
		}
		if chunk <= 0 {
			chunk = 1
		}
		pcm = append(pcm, opl.GenerateStereoS16(chunk)...)
		remaining -= chunk
		framesUntilNext -= chunk
	}
	return pcm, nil
}

func buildNukedSeqDumpTool() (string, error) {
	nukedToolOnce.Do(func() {
		out := filepath.Join(os.TempDir(), "nuked-opl3-seq-dump-compare-wav")
		cmd := exec.Command(
			"gcc",
			"-O2",
			"-I", "third_party/nuked-opl3",
			"-o", out,
			"bench/nuked_opl3_seq_dump_raw.c",
			"third_party/nuked-opl3/opl3.c",
			"-lm",
		)
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		if err != nil {
			nukedToolErr = fmt.Errorf("gcc failed: %w: %s", err, strings.TrimSpace(string(output)))
			return
		}
		nukedToolPath = out
	})
	if nukedToolErr != nil {
		return "", nukedToolErr
	}
	return nukedToolPath, nil
}

func renderNukedSeq(sampleRate, tickRate, frames int, seqPath string) ([]int16, error) {
	tool, err := buildNukedSeqDumpTool()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(tool, strconv.Itoa(sampleRate), strconv.Itoa(tickRate), strconv.Itoa(frames), seqPath)
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(output) != frames*4 {
		return nil, fmt.Errorf("unexpected raw nuked output size=%d want=%d", len(output), frames*4)
	}
	pcm := make([]int16, frames*2)
	for i := 0; i < len(pcm); i++ {
		pcm[i] = int16(binary.LittleEndian.Uint16(output[i*2 : i*2+2]))
	}
	return pcm, nil
}

func writeStereoS16WAV(path string, sampleRate int, pcm []int16) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dataSize := uint32(len(pcm) * 2)
	byteRate := uint32(sampleRate * 4)
	blockAlign := uint16(4)
	riffSize := uint32(36) + dataSize

	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, riffSize); err != nil {
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
	if err := binary.Write(f, binary.LittleEndian, uint16(2)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(16)); err != nil {
		return err
	}
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil {
		return err
	}
	for _, sample := range pcm {
		if err := binary.Write(f, binary.LittleEndian, sample); err != nil {
			return err
		}
	}
	return nil
}

func runtimeWorkers() int {
	return max(1, runtime.NumCPU())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
