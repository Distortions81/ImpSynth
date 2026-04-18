package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	impsynth "github.com/Distortions81/impsynth"
	"github.com/remeh/sizedwaitgroup"
)

const (
	sampleRate       = 49716
	compareChunkFrames = 2048
	wolfTickRate     = 700
	doomTickRate     = 140
	minChunkEnergy   = 64
	minSpecChunkEnergy = 4096
	minSpecChunkPeak   = 1
)

type manifestSong struct {
	Name  string `json:"name,omitempty"`
	Lump  string `json:"lump,omitempty"`
	File  string `json:"file"`
	Count int    `json:"event_count"`
}

type manifest struct {
	Source  string         `json:"source"`
	TicRate int            `json:"tic_rate,omitempty"`
	Songs   []manifestSong `json:"songs"`
}

type channelMetric struct {
	Channel     int
	Spec        float64
	EnergyRatio float64
	MaxDelta    int
	GotEnergy   int64
	WantEnergy  int64
}

type songMetric struct {
	Label        string
	Path         string
	Frames       int
	ActiveCount  int
	Channels     []channelMetric
	MeanSpec     float64
	MinSpec      float64
	MaxEnergyDev float64
}

type corpusMetric struct {
	Title        string
	Source       string
	Channels     int
	TickRate     int
	Songs        []songMetric
	WorstSpec    songMetric
	WorstChannel channelMetric
}

type corpusConfig struct {
	title           string
	manifestPath    string
	baseDir         string
	defaultTickRate int
	channels        int
	newSynth        func(int) *impsynth.Synth
}

type songPlan struct {
	corpusSlug  string
	corpusName  string
	label       string
	path        string
	tickRate    int
	frames      int
	channels    []int
	seqPaths    map[int]string
}

type renderJob struct {
	songIndex int
	channel   int
	renderer  string
	seqPath   string
	cachePath string
	tickRate  int
	frames    int
	newSynth  func(int) *impsynth.Synth
}

type renderResult struct {
	songIndex int
	channel   int
	renderer  string
	path      string
}

var (
	nukedToolOnce  sync.Once
	nukedToolPath  string
	nukedToolErr   error
	renderCacheDir string
)

func main() {
	var outPath string
	flag.StringVar(&outPath, "out", filepath.Join("docs", "shareware-channel-findings.md"), "output markdown path")
	flag.StringVar(&renderCacheDir, "render-cache", filepath.Join("testdata", "shareware-render-cache"), "directory for run-scoped render PCM data")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fail(err)
	}
	if renderCacheDir != "" {
		if err := os.RemoveAll(renderCacheDir); err != nil {
			fail(err)
		}
		if err := os.MkdirAll(renderCacheDir, 0o755); err != nil {
			fail(err)
		}
	}

	configs := []corpusConfig{
		{
			title:           "Wolf3D Shareware",
			manifestPath:    filepath.Join("testdata", "wolf3d-shareware-music", "manifest.json"),
			baseDir:         filepath.Join("testdata", "wolf3d-shareware-music"),
			defaultTickRate: wolfTickRate,
			channels:        9,
			newSynth:        func(rate int) *impsynth.Synth { return impsynth.NewOPL2(rate) },
		},
		{
			title:           "DOOM Shareware",
			manifestPath:    filepath.Join("testdata", "doom-shareware-music", "manifest.json"),
			baseDir:         filepath.Join("testdata", "doom-shareware-music"),
			defaultTickRate: doomTickRate,
			channels:        18,
			newSynth:        func(rate int) *impsynth.Synth { return impsynth.New(rate) },
		},
	}

	corpora := make([]corpusMetric, 0, len(configs))
	for _, cfg := range configs {
		corpus, err := analyzeCorpus(cfg)
		if err != nil {
			fail(err)
		}
		corpora = append(corpora, corpus)
		if err := writeMarkdown(outPath, corpora...); err != nil {
			fail(err)
		}
	}

	if err := writeMarkdown(outPath, corpora...); err != nil {
		fail(err)
	}
	fmt.Println(outPath)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func analyzeCorpus(cfg corpusConfig) (corpusMetric, error) {
	m, err := loadManifest(cfg.manifestPath)
	if err != nil {
		return corpusMetric{}, err
	}
	tickRate := m.TicRate
	if tickRate <= 0 {
		tickRate = cfg.defaultTickRate
	}

	corpusSlug := safeSlug(filepath.Base(cfg.baseDir))
	out := corpusMetric{
		Title:    cfg.title,
		Source:   m.Source,
		Channels: cfg.channels,
		TickRate: tickRate,
		Songs:    make([]songMetric, 0, len(m.Songs)),
	}

	totalWorkers := runtime.NumCPU()
	if totalWorkers < 1 {
		totalWorkers = 1
	}
	songWorkers := songParallelism(totalWorkers, len(m.Songs))
	perSongWorkers := totalWorkers / songWorkers
	if perSongWorkers < 1 {
		perSongWorkers = 1
	}
	metrics := make([]songMetric, len(m.Songs))
	errCh := make(chan error, len(m.Songs))
	swg := sizedwaitgroup.New(songWorkers)

	for songIndex, song := range m.Songs {
		songIndex := songIndex
		song := song
		swg.Add()
		go func() {
			defer swg.Done()
			plan, jobs, cleanup, err := prepareSongPlan(corpusSlug, cfg, tickRate, songIndex, song)
			if err != nil {
				errCh <- err
				return
			}
			results, err := runRenderJobs(jobs, perSongWorkers)
			cleanup()
			if err != nil {
				errCh <- err
				return
			}
			metric, err := compareSong(plan, songIndex, results, perSongWorkers)
			if err != nil {
				errCh <- err
				return
			}
			metrics[songIndex] = metric
		}()
	}
	swg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return corpusMetric{}, err
		}
	}
	for _, metric := range metrics {
		out.Songs = append(out.Songs, metric)
	}

	finalizeCorpus(&out)
	return out, nil
}

func songParallelism(totalWorkers, songCount int) int {
	if songCount < 1 {
		return 1
	}
	if totalWorkers < 8 || songCount == 1 {
		return 1
	}
	n := totalWorkers / 12
	if n < 2 {
		n = 2
	}
	if n > songCount {
		n = songCount
	}
	return n
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

func prepareSongPlan(corpusSlug string, cfg corpusConfig, tickRate int, songIndex int, song manifestSong) (songPlan, []renderJob, func(), error) {
	jobs := make([]renderJob, 0, cfg.channels*2)
	var tempFiles []string
	cleanup := func() {
		for _, path := range tempFiles {
			_ = os.Remove(path)
		}
	}

	label := strings.TrimSpace(song.Name)
	if label == "" {
		label = strings.TrimSpace(song.Lump)
	}
	if label == "" {
		label = strings.TrimSuffix(filepath.Base(song.File), filepath.Ext(song.File))
	}
	path := filepath.Join(cfg.baseDir, song.File)
	totalFrames, err := seqTotalFrames(path, tickRate)
	if err != nil {
		cleanup()
		return songPlan{}, nil, nil, fmt.Errorf("duration for %s: %w", label, err)
	}
	channels, err := detectActiveChannels(path, cfg.channels)
	if err != nil {
		cleanup()
		return songPlan{}, nil, nil, fmt.Errorf("detect active channels for %s: %w", label, err)
	}
	seqPaths := make(map[int]string, len(channels))
	for _, ch := range channels {
		tmp, err := filterSeqForChannel(path, ch)
		if err != nil {
			cleanup()
			return songPlan{}, nil, nil, fmt.Errorf("filter %s ch%d: %w", label, ch, err)
		}
		tempFiles = append(tempFiles, tmp)
		seqPaths[ch] = tmp
		jobs = append(jobs,
			renderJob{
				songIndex: songIndex,
				channel:   ch,
				renderer:  "impsynth",
				seqPath:   tmp,
				cachePath: renderPCMPath("impsynth", corpusSlug, path, tickRate, ch, totalFrames),
				tickRate:  tickRate,
				frames:    totalFrames,
				newSynth:  cfg.newSynth,
			},
			renderJob{
				songIndex: songIndex,
				channel:   ch,
				renderer:  "nuked",
				seqPath:   tmp,
				cachePath: renderPCMPath("nuked", corpusSlug, path, tickRate, ch, totalFrames),
				tickRate:  tickRate,
				frames:    totalFrames,
			},
		)
	}

	return songPlan{
		corpusSlug: corpusSlug,
		corpusName: cfg.title,
		label:      label,
		path:       path,
		tickRate:   tickRate,
		frames:     totalFrames,
		channels:   channels,
		seqPaths:   seqPaths,
	}, jobs, cleanup, nil
}

func runRenderJobs(jobs []renderJob, limit int) (map[int]map[int]map[string]string, error) {
	if limit < 1 {
		limit = 1
	}
	swg := sizedwaitgroup.New(limit)
	errCh := make(chan error, len(jobs))
	results := make(map[int]map[int]map[string]string)
	var mu sync.Mutex

	for _, job := range jobs {
		job := job
		swg.Add()
		go func() {
			defer swg.Done()
			path, err := runRenderJob(job)
			if err != nil {
				errCh <- fmt.Errorf("%s render song=%d ch=%d: %w", job.renderer, job.songIndex, job.channel, err)
				return
			}
			mu.Lock()
			bySong := results[job.songIndex]
			if bySong == nil {
				bySong = make(map[int]map[string]string)
				results[job.songIndex] = bySong
			}
			byChannel := bySong[job.channel]
			if byChannel == nil {
				byChannel = make(map[string]string, 2)
				bySong[job.channel] = byChannel
			}
			byChannel[job.renderer] = path
			mu.Unlock()
		}()
	}

	swg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func compareSong(plan songPlan, songIndex int, rendered map[int]map[int]map[string]string, limit int) (songMetric, error) {
	out := songMetric{
		Label:    plan.label,
		Path:     plan.path,
		Frames:   plan.frames,
		Channels: make([]channelMetric, 0, len(plan.channels)),
		MinSpec:  1,
	}

	bySong := rendered[songIndex]
	if limit < 1 {
		limit = 1
	}
	swg := sizedwaitgroup.New(limit)
	errCh := make(chan error, len(plan.channels))
	metricCh := make(chan channelMetric, len(plan.channels))

	for _, ch := range plan.channels {
		ch := ch
		swg.Add()
		go func() {
			defer swg.Done()
			pair := bySong[ch]
			if pair == nil {
				return
			}
			gotPath := pair["impsynth"]
			wantPath := pair["nuked"]
			if gotPath == "" || wantPath == "" {
				return
			}
			got, err := readPCMFile(gotPath)
			if err != nil {
				errCh <- fmt.Errorf("read impsynth pcm for %s ch%d: %w", plan.label, ch, err)
				return
			}
			want, err := readPCMFile(wantPath)
			if err != nil {
				errCh <- fmt.Errorf("read nuked pcm for %s ch%d: %w", plan.label, ch, err)
				return
			}
			spec, gotEnergy, wantEnergy, maxDelta := compareWholeSongPCM(got, want)
			if gotEnergy < minChunkEnergy && wantEnergy < minChunkEnergy {
				return
			}
			ratio := 0.0
			if wantEnergy > 0 {
				ratio = float64(gotEnergy) / float64(wantEnergy)
			}
			metricCh <- channelMetric{
				Channel:     ch,
				Spec:        spec,
				EnergyRatio: ratio,
				MaxDelta:    maxDelta,
				GotEnergy:   gotEnergy,
				WantEnergy:  wantEnergy,
			}
		}()
	}
	swg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return songMetric{}, err
		}
	}
	close(metricCh)
	for metric := range metricCh {
		out.Channels = append(out.Channels, metric)
	}

	for _, metric := range out.Channels {
		out.MeanSpec += metric.Spec
		if metric.Spec < out.MinSpec {
			out.MinSpec = metric.Spec
		}
		dev := math.Abs(1 - metric.EnergyRatio)
		if metric.EnergyRatio == 0 && metric.WantEnergy == 0 {
			dev = 0
		}
		if dev > out.MaxEnergyDev {
			out.MaxEnergyDev = dev
		}
	}
	out.ActiveCount = len(out.Channels)
	sort.Slice(out.Channels, func(i, j int) bool {
		if out.Channels[i].Spec == out.Channels[j].Spec {
			return out.Channels[i].Channel < out.Channels[j].Channel
		}
		return out.Channels[i].Spec < out.Channels[j].Spec
	})
	if len(out.Channels) > 0 {
		out.MeanSpec /= float64(len(out.Channels))
	} else {
		out.MinSpec = 1
	}

	return out, nil
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

func runRenderJob(job renderJob) (string, error) {
	if job.cachePath != "" {
		if info, err := os.Stat(job.cachePath); err == nil && info.Size() == int64(job.frames*4) {
			return job.cachePath, nil
		}
	}

	var pcm []int16
	var err error
	switch job.renderer {
	case "impsynth":
		pcm, err = renderImpSynthSeq(sampleRate, job.tickRate, job.frames, job.seqPath, job.newSynth)
	case "nuked":
		pcm, err = renderNukedSeq(sampleRate, job.tickRate, job.frames, job.seqPath)
	default:
		return "", fmt.Errorf("unknown renderer %q", job.renderer)
	}
	if err != nil {
		return "", err
	}
	if job.cachePath == "" {
		return "", fmt.Errorf("missing render path for %s", job.renderer)
	}
	if err := writePCMFile(job.cachePath, pcm); err != nil {
		return "", err
	}
	return job.cachePath, nil
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
				// This filtered stream has no delayed events left to advance time.
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
		out := filepath.Join(os.TempDir(), "nuked-opl3-seq-dump-analysis")
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

func readPCMFile(path string) ([]int16, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("pcm size %d is not a stereo int16 multiple", len(data))
	}
	pcm := make([]int16, len(data)/2)
	for i := 0; i < len(pcm); i++ {
		pcm[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return pcm, nil
}

func writePCMFile(path string, pcm []int16) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data := make([]byte, len(pcm)*2)
	for i, sample := range pcm {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(sample))
	}
	return os.WriteFile(path, data, 0o644)
}

func renderPCMPath(renderer, corpusSlug, songPath string, tickRate, channel, frames int) string {
	if renderCacheDir == "" {
		return ""
	}
	songSlug := safeSlug(strings.TrimSuffix(filepath.Base(songPath), filepath.Ext(songPath)))
	name := fmt.Sprintf("sr%d-tr%d-ch%02d-fr%d.pcm", sampleRate, tickRate, channel, frames)
	return filepath.Join(renderCacheDir, renderer, corpusSlug, songSlug, name)
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

func detectActiveChannels(path string, channels int) ([]int, error) {
	events, err := readSeq(path)
	if err != nil {
		return nil, err
	}
	active := make([]bool, channels)
	for _, ev := range events {
		for ch := 0; ch < channels; ch++ {
			if isChannelSpecificReg(ev.Reg, ch) {
				active[ch] = true
			}
		}
	}
	out := make([]int, 0, channels)
	for ch, ok := range active {
		if ok {
			out = append(out, ch)
		}
	}
	return out, nil
}

func filterSeqForChannel(src string, ch int) (string, error) {
	events, err := readSeq(src)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "impsynth-shareware-ch-*.seq")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	filtered := make([]seqEvent, 0, len(events))
	pendingDelay := uint32(0)
	for _, ev := range events {
		if keepRegForChannel(ev.Reg, ch) {
			delay := uint32(ev.Delay)
			if len(filtered) > 0 {
				carried := pendingDelay
				if carried > 0xffff {
					carried = 0xffff
				}
				filtered[len(filtered)-1].Delay = uint16(carried)
			}
			filtered = append(filtered, seqEvent{
				Reg:   ev.Reg,
				Value: ev.Value,
				Delay: uint16(delay),
			})
			pendingDelay = delay
		} else {
			pendingDelay += uint32(ev.Delay)
			if pendingDelay > 0xffff {
				pendingDelay = 0xffff
			}
		}
	}
	if len(filtered) > 0 {
		carried := pendingDelay
		if carried > 0xffff {
			carried = 0xffff
		}
		filtered[len(filtered)-1].Delay = uint16(carried)
	}
	buf := make([]byte, 0, len(filtered)*5)
	for _, ev := range filtered {
		buf = binary.LittleEndian.AppendUint16(buf, ev.Reg)
		buf = append(buf, ev.Value)
		buf = binary.LittleEndian.AppendUint16(buf, ev.Delay)
	}
	if _, err := tmp.Write(buf); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func keepRegForChannel(reg uint16, ch int) bool {
	if reg == 0x01 || reg == 0x08 || reg == 0xBD || reg == 0x104 || reg == 0x105 {
		return true
	}
	return isChannelSpecificReg(reg, ch)
}

func isChannelSpecificReg(reg uint16, ch int) bool {
	base := uint16(0)
	localCh := ch
	if ch >= 9 {
		base = 0x100
		localCh = ch - 9
	}
	switch {
	case reg >= base+0xA0 && reg <= base+0xA8:
		return int(reg-(base+0xA0)) == localCh
	case reg >= base+0xB0 && reg <= base+0xB8:
		return int(reg-(base+0xB0)) == localCh
	case reg >= base+0xC0 && reg <= base+0xC8:
		return int(reg-(base+0xC0)) == localCh
	case reg >= base+0xD0 && reg <= base+0xD8:
		return int(reg-(base+0xD0)) == localCh
	case isSlotRegister(base, reg, 0x20, localCh):
		return true
	case isSlotRegister(base, reg, 0x40, localCh):
		return true
	case isSlotRegister(base, reg, 0x60, localCh):
		return true
	case isSlotRegister(base, reg, 0x80, localCh):
		return true
	case isSlotRegister(base, reg, 0xE0, localCh):
		return true
	default:
		return false
	}
}

func isSlotRegister(base, reg uint16, prefix uint16, ch int) bool {
	slots := [9][2]uint16{
		{0, 3}, {1, 4}, {2, 5},
		{8, 11}, {9, 12}, {10, 13},
		{16, 19}, {17, 20}, {18, 21},
	}
	if ch < 0 || ch >= len(slots) {
		return false
	}
	mod := base + prefix + slots[ch][0]
	car := base + prefix + slots[ch][1]
	return reg == mod || reg == car
}

func sliceStereoFrames(pcm []int16, frameOffset int, frameCount int) []int16 {
	start := frameOffset * 2
	if start > len(pcm) {
		start = len(pcm)
	}
	end := start + frameCount*2
	if end > len(pcm) {
		end = len(pcm)
	}
	out := make([]int16, end-start)
	copy(out, pcm[start:end])
	return out
}

func chunkStats(pcm []int16) (int64, int) {
	var energy int64
	peak := 0
	for i := 0; i+1 < len(pcm); i += 2 {
		s := (int(pcm[i]) + int(pcm[i+1])) / 2
		if s < 0 {
			s = -s
		}
		energy += int64(s)
		if s > peak {
			peak = s
		}
	}
	return energy, peak
}

func compareWholeSongPCM(got, want []int16) (float64, int64, int64, int) {
	gotEnergy := monoAbsEnergy(got)
	wantEnergy := monoAbsEnergy(want)
	maxDelta := maxPCMDelta(got, want)
	frames := len(got) / 2
	if other := len(want) / 2; other < frames {
		frames = other
	}
	if frames <= 0 {
		return 0, gotEnergy, wantEnergy, maxDelta
	}

	specSum := 0.0
	specCount := 0
	for start := 0; start < frames; start += compareChunkFrames {
		chunkFrames := compareChunkFrames
		if remain := frames - start; remain < chunkFrames {
			chunkFrames = remain
		}
		gotChunk := sliceStereoFrames(got, start, chunkFrames)
		wantChunk := sliceStereoFrames(want, start, chunkFrames)
		gotChunkEnergy, gotPeak := chunkStats(gotChunk)
		wantChunkEnergy, wantPeak := chunkStats(wantChunk)
		if (gotChunkEnergy < minSpecChunkEnergy && wantChunkEnergy < minSpecChunkEnergy) ||
			(gotPeak <= minSpecChunkPeak && wantPeak <= minSpecChunkPeak) {
			continue
		}
		specSum += spectrumCosineSimilarity(gotChunk, wantChunk, 512)
		specCount++
	}
	if specCount == 0 {
		return 0, gotEnergy, wantEnergy, maxDelta
	}
	return specSum / float64(specCount), gotEnergy, wantEnergy, maxDelta
}

func maxPCMDelta(a, b []int16) int {
	if len(a) != len(b) {
		return int(^uint(0) >> 1)
	}
	max := 0
	for i := range a {
		d := int(a[i]) - int(b[i])
		if d < 0 {
			d = -d
		}
		if d > max {
			max = d
		}
	}
	return max
}

func monoAbsEnergy(pcm []int16) int64 {
	var total int64
	for i := 0; i+1 < len(pcm); i += 2 {
		s := int64(pcm[i]+pcm[i+1]) / 2
		if s < 0 {
			s = -s
		}
		total += s
	}
	return total
}

func monoFrames(pcm []int16) []float64 {
	out := make([]float64, 0, len(pcm)/2)
	for i := 0; i+1 < len(pcm); i += 2 {
		out = append(out, float64(int(pcm[i])+int(pcm[i+1]))/2.0)
	}
	return out
}

func normalizedMagnitudeSpectrum(pcm []int16, fftSize int) []float64 {
	frames := monoFrames(pcm)
	if len(frames) < fftSize {
		fftSize = len(frames)
	}
	if fftSize < 8 {
		return nil
	}
	windowed := make([]float64, fftSize)
	for i := 0; i < fftSize; i++ {
		w := 0.5 - 0.5*math.Cos((2*math.Pi*float64(i))/float64(fftSize-1))
		windowed[i] = frames[i] * w
	}
	bins := fftSize/2 + 1
	spec := make([]float64, bins)
	var sum float64
	for k := 0; k < bins; k++ {
		var re, im float64
		for n := 0; n < fftSize; n++ {
			phase := -2 * math.Pi * float64(k*n) / float64(fftSize)
			re += windowed[n] * math.Cos(phase)
			im += windowed[n] * math.Sin(phase)
		}
		mag := math.Hypot(re, im)
		spec[k] = mag
		sum += mag
	}
	if sum > 0 {
		for i := range spec {
			spec[i] /= sum
		}
	}
	return spec
}

func spectrumCosineSimilarity(a, b []int16, fftSize int) float64 {
	sa := normalizedMagnitudeSpectrum(a, fftSize)
	sb := normalizedMagnitudeSpectrum(b, fftSize)
	if len(sa) == 0 || len(sb) == 0 || len(sa) != len(sb) {
		return 0
	}
	var dot, na, nb float64
	for i := range sa {
		dot += sa[i] * sb[i]
		na += sa[i] * sa[i]
		nb += sb[i] * sb[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / math.Sqrt(na*nb)
}

func renderMarkdown(corpora ...corpusMetric) string {
	var b strings.Builder
	b.WriteString("# Shareware Channel Comparison\n\n")
	b.WriteString("This document compares per-channel `ImpSynth` output against `Nuked-OPL3` using the exported Wolf3D and DOOM shareware sequence fixtures.\n\n")
	b.WriteString("Method:\n")
	b.WriteString("- Render one full song cycle for each active OPL voice/channel in `ImpSynth` and `Nuked-OPL3`.\n")
	b.WriteString("- Split the rendered PCM into in-memory chunks for spectral comparison.\n")
	b.WriteString("- Aggregate chunk similarities plus full-song energy ratio and max sample delta.\n\n")
	for _, corpus := range corpora {
		fmt.Fprintf(&b, "## %s\n\n", corpus.Title)
		fmt.Fprintf(&b, "Source: `%s`\n", corpus.Source)
		fmt.Fprintf(&b, "Tick rate: `%d`\n", corpus.TickRate)
		fmt.Fprintf(&b, "Channels analyzed: `%d`\n\n", corpus.Channels)
		if len(corpus.Songs) == 0 {
			b.WriteString("No songs analyzed.\n\n")
			continue
		}
		fmt.Fprintf(&b, "Worst observed channel: `%s` channel `%d` with spectral similarity `%.3f`, energy ratio `%.3fx`, max delta `%d`.\n\n",
			corpus.WorstSpec.Label, corpus.WorstChannel.Channel, corpus.WorstChannel.Spec, corpus.WorstChannel.EnergyRatio, corpus.WorstChannel.MaxDelta)
		b.WriteString("| Song | Frames | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |\n")
		b.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
		for _, song := range corpus.Songs {
			if len(song.Channels) == 0 {
				continue
			}
			worst := song.Channels[0]
			fmt.Fprintf(&b, "| %s | %d | %d | %.3f | %d | %.3f | %.3fx |\n",
				song.Label, song.Frames, song.ActiveCount, song.MeanSpec, worst.Channel, worst.Spec, worst.EnergyRatio)
		}
		b.WriteString("\n")
		for _, song := range corpus.Songs {
			if len(song.Channels) == 0 {
				continue
			}
			b.WriteString("### ")
			b.WriteString(song.Label)
			b.WriteString("\n\n")
			fmt.Fprintf(&b, "Full-song frames: `%d`. Active channels: `%d`. Worst channels first.\n\n", song.Frames, song.ActiveCount)
			b.WriteString("| Channel | Spec | Energy Ratio | Max Delta |\n")
			b.WriteString("| --- | ---: | ---: | ---: |\n")
			limit := len(song.Channels)
			if limit > 6 {
				limit = 6
			}
			for i := 0; i < limit; i++ {
				ch := song.Channels[i]
				fmt.Fprintf(&b, "| %d | %.3f | %.3fx | %d |\n", ch.Channel, ch.Spec, ch.EnergyRatio, ch.MaxDelta)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func writeMarkdown(path string, corpora ...corpusMetric) error {
	return os.WriteFile(path, []byte(renderMarkdown(corpora...)), 0o644)
}

func finalizeCorpus(corpus *corpusMetric) {
	if corpus == nil {
		return
	}
	sort.Slice(corpus.Songs, func(i, j int) bool {
		if corpus.Songs[i].MinSpec == corpus.Songs[j].MinSpec {
			return corpus.Songs[i].Label < corpus.Songs[j].Label
		}
		return corpus.Songs[i].MinSpec < corpus.Songs[j].MinSpec
	})
	if len(corpus.Songs) == 0 {
		corpus.WorstSpec = songMetric{}
		corpus.WorstChannel = channelMetric{}
		return
	}
	found := false
	for _, song := range corpus.Songs {
		for _, ch := range song.Channels {
			if !found || ch.Spec < corpus.WorstChannel.Spec {
				corpus.WorstSpec = song
				corpus.WorstChannel = ch
				found = true
			}
		}
	}
	if !found {
		corpus.WorstSpec = corpus.Songs[0]
		corpus.WorstChannel = channelMetric{}
	}
}

func cloneCorpus(in corpusMetric) corpusMetric {
	out := in
	out.Songs = append([]songMetric(nil), in.Songs...)
	for i := range out.Songs {
		out.Songs[i].Channels = append([]channelMetric(nil), in.Songs[i].Channels...)
	}
	if len(out.WorstSpec.Channels) > 0 {
		out.WorstSpec.Channels = append([]channelMetric(nil), out.WorstSpec.Channels...)
	}
	return out
}

func safeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
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
		return "unnamed"
	}
	return out
}
