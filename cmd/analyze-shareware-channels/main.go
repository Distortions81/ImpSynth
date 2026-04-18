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
	sampleRate      = 49716
	windowFrames    = 2048
	chunkFrames     = 512
	wolfTickRate    = 700
	doomTickRate    = 140
	minWindowEnergy = 64
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
	Skip         int
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

type renderedChannel struct {
	channel  int
	renderer string
	chunks   [][]int16
}

type chunkMetric struct {
	channel    int
	spec       float64
	gotEnergy  int64
	wantEnergy int64
	maxDelta   int
	weight     float64
}

type channelAccumulator struct {
	specWeighted float64
	weightTotal  float64
	gotEnergy    int64
	wantEnergy   int64
	maxDelta     int
}

type chunkKey struct {
	channel int
	chunk   int
}

type renderJob struct {
	key       chunkKey
	seqPath   string
	cachePath string
	tickRate  int
	start     int
	frames    int
	renderer  string
	newSynth  func(int) *impsynth.Synth
}

type renderResult struct {
	key      chunkKey
	renderer string
	pcm      []int16
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
	flag.StringVar(&renderCacheDir, "render-cache", filepath.Join("testdata", "shareware-render-cache"), "directory for reusable render cache data")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fail(err)
	}
	if renderCacheDir != "" {
		if err := os.MkdirAll(renderCacheDir, 0o755); err != nil {
			fail(err)
		}
	}

	var wolf corpusMetric
	var doom corpusMetric
	writeProgress := func() error {
		return writeMarkdown(outPath, wolf, doom)
	}

	wolf, err := analyzeCorpus(
		"Wolf3D Shareware",
		filepath.Join("testdata", "wolf3d-shareware-music", "manifest.json"),
		filepath.Join("testdata", "wolf3d-shareware-music"),
		wolfTickRate,
		9,
		func(rate int) *impsynth.Synth { return impsynth.NewOPL2(rate) },
		func(progress corpusMetric) error {
			wolf = progress
			return writeProgress()
		},
	)
	if err != nil {
		fail(err)
	}
	doom, err = analyzeCorpus(
		"DOOM Shareware",
		filepath.Join("testdata", "doom-shareware-music", "manifest.json"),
		filepath.Join("testdata", "doom-shareware-music"),
		doomTickRate,
		18,
		func(rate int) *impsynth.Synth { return impsynth.New(rate) },
		func(progress corpusMetric) error {
			doom = progress
			return writeProgress()
		},
	)
	if err != nil {
		fail(err)
	}

	if err := writeMarkdown(outPath, wolf, doom); err != nil {
		fail(err)
	}
	fmt.Println(outPath)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
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

func analyzeCorpus(title, manifestPath, baseDir string, defaultTickRate, channels int, newSynth func(int) *impsynth.Synth, progress func(corpusMetric) error) (corpusMetric, error) {
	var m manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return corpusMetric{}, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return corpusMetric{}, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}
	tickRate := m.TicRate
	if tickRate <= 0 {
		tickRate = defaultTickRate
	}
	out := corpusMetric{
		Title:    title,
		Source:   m.Source,
		Channels: channels,
		TickRate: tickRate,
		Songs:    make([]songMetric, 0, len(m.Songs)),
	}
	limit := runtime.NumCPU()
	if limit < 1 {
		limit = 1
	}
	swg := sizedwaitgroup.New(limit)
	errCh := make(chan error, len(m.Songs))
	var mu sync.Mutex
	for _, song := range m.Songs {
		song := song
		swg.Add()
		go func() {
			defer swg.Done()
			label := strings.TrimSpace(song.Name)
			if label == "" {
				label = strings.TrimSpace(song.Lump)
			}
			path := filepath.Join(baseDir, song.File)
			metric, err := analyzeSong(safeSlug(filepath.Base(baseDir)), label, path, tickRate, channels, newSynth)
			if err != nil {
				errCh <- err
				return
			}
			mu.Lock()
			out.Songs = append(out.Songs, metric)
			finalizeCorpus(&out)
			snapshot := cloneCorpus(out)
			mu.Unlock()
			if progress != nil {
				if err := progress(snapshot); err != nil {
					errCh <- err
				}
			}
		}()
	}
	swg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return corpusMetric{}, err
		}
	}
	finalizeCorpus(&out)
	return out, nil
}

func analyzeSong(corpusSlug, label, path string, tickRate, channels int, newSynth func(int) *impsynth.Synth) (songMetric, error) {
	windowStart, err := findWindow(path, tickRate, newSynth)
	if err != nil {
		return songMetric{}, fmt.Errorf("find window for %s: %w", label, err)
	}
	active, err := detectActiveChannels(path, channels)
	if err != nil {
		return songMetric{}, fmt.Errorf("detect active channels for %s: %w", label, err)
	}
	out := songMetric{
		Label:    label,
		Path:     path,
		Skip:     windowStart,
		Frames:   windowFrames,
		Channels: make([]channelMetric, 0, len(active)),
		MinSpec:  1,
	}
	tempSeqs := make(map[int]string, len(active))
	defer func() {
		for _, p := range tempSeqs {
			_ = os.Remove(p)
		}
	}()
	for _, ch := range active {
		tmp, err := filterSeqForChannel(path, ch)
		if err != nil {
			return songMetric{}, fmt.Errorf("filter %s ch%d: %w", label, ch, err)
		}
		tempSeqs[ch] = tmp
	}

	totalChunks := (windowFrames + chunkFrames - 1) / chunkFrames
	jobs := make([]renderJob, 0, len(active)*totalChunks*2)
	for _, ch := range active {
		for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
			start := windowStart + chunkIndex*chunkFrames
			frames := min(chunkFrames, windowFrames-chunkIndex*chunkFrames)
			key := chunkKey{channel: ch, chunk: chunkIndex}
			jobs = append(jobs,
				renderJob{
					key:       key,
					seqPath:   tempSeqs[ch],
					cachePath: "",
					tickRate:  tickRate,
					start:     start,
					frames:    frames,
					renderer:  "impsynth",
					newSynth:  newSynth,
				},
				renderJob{
					key:       key,
					seqPath:   tempSeqs[ch],
					cachePath: nukedCachePath(corpusSlug, path, tickRate, key, start, frames),
					tickRate:  tickRate,
					start:     start,
					frames:    frames,
					renderer:  "nuked",
				},
			)
		}
	}

	rendered := make(map[chunkKey]map[string][]int16, len(jobs))
	var renderMu sync.Mutex
	renderErrCh := make(chan error, len(jobs))
	limit := runtime.NumCPU()
	if limit < 1 {
		limit = 1
	}
	renderWG := sizedwaitgroup.New(limit)
	for _, job := range jobs {
		job := job
		renderWG.Add()
		go func() {
			defer renderWG.Done()
			pcm, err := runRenderJob(job)
			if err != nil {
				renderErrCh <- fmt.Errorf("%s render %s ch%d chunk%d: %w", job.renderer, label, job.key.channel, job.key.chunk, err)
				return
			}
			renderMu.Lock()
			entry := rendered[job.key]
			if entry == nil {
				entry = make(map[string][]int16, 2)
				rendered[job.key] = entry
			}
			entry[job.renderer] = pcm
			renderMu.Unlock()
		}()
	}
	renderWG.Wait()
	close(renderErrCh)
	for err := range renderErrCh {
		if err != nil {
			return songMetric{}, err
		}
	}

	accumulators := make(map[int]*channelAccumulator, len(active))
	var compareMu sync.Mutex
	compareErrCh := make(chan error, len(rendered))
	compareWG := sizedwaitgroup.New(limit)
	for key, entry := range rendered {
		gotChunk, okGot := entry["impsynth"]
		wantChunk, okWant := entry["nuked"]
		if !okGot || !okWant {
			continue
		}
		key := key
		gotPCM := gotChunk
		wantPCM := wantChunk
		compareWG.Add()
		go func() {
			defer compareWG.Done()
			gotEnergy := monoAbsEnergy(gotPCM)
			wantEnergy := monoAbsEnergy(wantPCM)
			if gotEnergy < minWindowEnergy && wantEnergy < minWindowEnergy {
				return
			}
			spec := spectrumCosineSimilarity(gotPCM, wantPCM, min(512, len(gotPCM)/2))
			weight := float64(gotEnergy + wantEnergy)
			if weight <= 0 {
				weight = 1
			}
			delta := maxPCMDelta(gotPCM, wantPCM)
			compareMu.Lock()
			acc := accumulators[key.channel]
			if acc == nil {
				acc = &channelAccumulator{}
				accumulators[key.channel] = acc
			}
			acc.specWeighted += spec * weight
			acc.weightTotal += weight
			acc.gotEnergy += gotEnergy
			acc.wantEnergy += wantEnergy
			if delta > acc.maxDelta {
				acc.maxDelta = delta
			}
			compareMu.Unlock()
		}()
	}
	compareWG.Wait()
	close(compareErrCh)
	for err := range compareErrCh {
		if err != nil {
			return songMetric{}, err
		}
	}

	for _, ch := range active {
		acc := accumulators[ch]
		if acc == nil || acc.weightTotal == 0 {
			continue
		}
		gotEnergy := acc.gotEnergy
		wantEnergy := acc.wantEnergy
		spec := acc.specWeighted / acc.weightTotal
		ratio := 0.0
		if wantEnergy > 0 {
			ratio = float64(gotEnergy) / float64(wantEnergy)
		}
		out.Channels = append(out.Channels, channelMetric{
			Channel:     ch,
			Spec:        spec,
			EnergyRatio: ratio,
			MaxDelta:    acc.maxDelta,
			GotEnergy:   gotEnergy,
			WantEnergy:  wantEnergy,
		})
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

func findWindow(seqPath string, tickRate int, newSynth func(int) *impsynth.Synth) (int, error) {
	const searchFrames = 16384
	full, err := renderImpSynthSeq(sampleRate, tickRate, searchFrames, seqPath, newSynth)
	if err != nil {
		return 0, err
	}
	bestStart := 0
	var bestEnergy int64 = -1
	for start := 0; start+windowFrames <= searchFrames; start += windowFrames / 2 {
		window := sliceStereoFrames(full, start, windowFrames)
		energy := monoAbsEnergy(window)
		if energy > bestEnergy {
			bestEnergy = energy
			bestStart = start
		}
	}
	return bestStart, nil
}

func runRenderJob(job renderJob) ([]int16, error) {
	totalFrames := job.start + job.frames
	switch job.renderer {
	case "impsynth":
		all, err := renderImpSynthSeq(sampleRate, job.tickRate, totalFrames, job.seqPath, job.newSynth)
		if err != nil {
			return nil, err
		}
		return sliceStereoFrames(all, job.start, job.frames), nil
	case "nuked":
		if job.cachePath != "" {
			if pcm, ok, err := readPCMCache(job.cachePath, job.frames); err != nil {
				return nil, err
			} else if ok {
				return pcm, nil
			}
		}
		all, err := renderNukedSeq(sampleRate, job.tickRate, totalFrames, job.seqPath)
		if err != nil {
			return nil, err
		}
		pcm := sliceStereoFrames(all, job.start, job.frames)
		if job.cachePath != "" {
			if err := writePCMCache(job.cachePath, pcm); err != nil {
				return nil, err
			}
		}
		return pcm, nil
	default:
		return nil, fmt.Errorf("unknown renderer %q", job.renderer)
	}
}

func nukedCachePath(corpusSlug, songPath string, tickRate int, key chunkKey, start, frames int) string {
	if renderCacheDir == "" {
		return ""
	}
	songSlug := safeSlug(strings.TrimSuffix(filepath.Base(songPath), filepath.Ext(songPath)))
	name := fmt.Sprintf("sr%d-tr%d-ch%02d-ck%02d-st%d-fr%d.pcm", sampleRate, tickRate, key.channel, key.chunk, start, frames)
	return filepath.Join(renderCacheDir, "nuked", corpusSlug, songSlug, name)
}

func readPCMCache(path string, frames int) ([]int16, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if len(data) != frames*4 {
		return nil, false, nil
	}
	pcm := make([]int16, frames*2)
	for i := 0; i < len(pcm); i++ {
		pcm[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return pcm, true, nil
}

func writePCMCache(path string, pcm []int16) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data := make([]byte, len(pcm)*2)
	for i, sample := range pcm {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(sample))
	}
	return os.WriteFile(path, data, 0o644)
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
		for framesUntilNext <= 0 {
			ev := events[eventIndex]
			eventIndex++
			if eventIndex >= len(events) {
				eventIndex = 0
			}
			opl.WriteReg(ev.Reg, ev.Value)
			framesUntilNext = int(ev.Delay) * tickFrames
			if framesUntilNext > 0 {
				break
			}
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
	buf := make([]byte, 0, len(events)*5)
	for _, ev := range events {
		if keepRegForChannel(ev.Reg, ch) {
			buf = binary.LittleEndian.AppendUint16(buf, ev.Reg)
			buf = append(buf, ev.Value)
			buf = binary.LittleEndian.AppendUint16(buf, ev.Delay)
		}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	b.WriteString("- Render a short high-energy window from each song.\n")
	b.WriteString("- Isolate each OPL channel by filtering its channel and operator register writes plus shared chip-global registers.\n")
	b.WriteString("- Compare `ImpSynth` and `Nuked-OPL3` with spectral cosine similarity, energy ratio, and max sample delta.\n\n")
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
		b.WriteString("| Song | Window Start | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |\n")
		b.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
		for _, song := range corpus.Songs {
			if len(song.Channels) == 0 {
				continue
			}
			worst := song.Channels[0]
			fmt.Fprintf(&b, "| %s | %d | %d | %.3f | %d | %.3f | %.3fx |\n",
				song.Label, song.Skip, song.ActiveCount, song.MeanSpec, worst.Channel, worst.Spec, worst.EnergyRatio)
		}
		b.WriteString("\n")
		for _, song := range corpus.Songs {
			if len(song.Channels) == 0 {
				continue
			}
			b.WriteString("### ")
			b.WriteString(song.Label)
			b.WriteString("\n\n")
			fmt.Fprintf(&b, "Window start: `%d` frames. Active channels: `%d`. Worst channels first.\n\n", song.Skip, song.ActiveCount)
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
