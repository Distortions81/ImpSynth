# ImpSynth

![ImpSynth logo](ImpSynth.png)

`impsynth` is a small Go OPL3-style FM synth library.

It focuses on the practical DMX/Doom-style register subset:
- 2-op voices
- operator envelopes
- feedback
- waveforms
- stereo pan
- OPL-style register writes

This implementation is based in part on the `Nuked-OPL3` OPL3 emulator by
Nuke.YKT, adapted here into a smaller Go library focused on a subset of OPL3
behavior.

## Install

```bash
go get github.com/Distortions81/impsynth
```

## Usage

```go
package main

import "github.com/Distortions81/impsynth"

func main() {
    opl := impsynth.New(49716)
    opl.Reset()
    opl.WriteReg(0x01, 0x20)
    opl.WriteReg(0x20, 0x21)
    opl.WriteReg(0x23, 0x01)
    opl.WriteReg(0x40, 0x08)
    opl.WriteReg(0x43, 0x00)
    opl.WriteReg(0x60, 0xF2)
    opl.WriteReg(0x63, 0xF2)
    opl.WriteReg(0x80, 0x24)
    opl.WriteReg(0x83, 0x24)
    opl.WriteReg(0xC0, 0x30)
    opl.WriteReg(0xA0, 0x98)
    opl.WriteReg(0xB0, 0x31)

    pcm := opl.GenerateStereoS16(2048)
    _ = pcm
}
```

## API

- `func New(sampleRate int) *Synth`
- `func (*Synth) Reset()`
- `func (*Synth) WriteReg(addr uint16, value uint8)`
- `func (*Synth) GenerateStereoS16(frames int) []int16`
- `func (*Synth) GenerateMonoU8(frames int) []byte`

## Example Program

This repo includes a small renderer that turns a simple melody CSV plus an OPL patch file into a `.wav`:

```bash
go run ./cmd/impsynth-wav-example examples/twinkle.csv twinkle.wav examples/patches/xylophone.json
```

Example render:

- [`impsynth-twinkle.mp3`](impsynth-twinkle.mp3)

The bundled [`examples/twinkle.csv`](examples/twinkle.csv) is the traditional French melody
`Ah! vous dirai-je, maman` (1761), which is in the public domain and is commonly used for
`Twinkle, Twinkle, Little Star`.

The bundled patch at [`examples/patches/xylophone.json`](examples/patches/xylophone.json)
defines a xylophone-like 2-op OPL voice using the same register values you would otherwise
write manually.

CSV format:

```text
start_ms,duration_ms,midi_note
0,500,60
500,500,60
1000,500,67
```

Patch format:

```json
{
  "name": "xylophone",
  "regs": {
    "0x20": 1,
    "0x23": 1,
    "0x40": 24
  }
}
```

## Benchmark

Benchmark coverage is limited to synth PCM generation and excludes WAV file writing.

For the cross-check against `Nuked-OPL3`, both implementations use the same
benchmark patch data:

- render size per operation: `2048` stereo `int16` frames
- register writes: `0x01=0x20`, `0x20=0x01`, `0x23=0x01`, `0x40=0x18`,
  `0x43=0x00`, `0x60=0xF4`, `0x63=0xF6`, `0x80=0x55`, `0x83=0x14`,
  `0xC0=0x30`, `0xA0=0x98`, `0xB0=0x31`

Commands:

```bash
go test -bench=BenchmarkGenerateStereoS16_2048Frames -benchmem -benchtime=2048x -run=^$ ./...
go test -bench=BenchmarkGenerateStereoS16_2048Frames_44100Hz -benchmem -benchtime=2048x -run=^$ ./...
./scripts/fetch-nuked-opl3.sh
./scripts/benchmark-nuked-opl3.sh 2048
./scripts/benchmark-nuked-opl3.sh 2048 44100
```

Result on March 13, 2026:

Native-rate (`49716 Hz`) result:

| Implementation | Benchmark | CPU | Iterations | Sample Rate | ns/op | MB/s |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `ImpSynth` | `BenchmarkGenerateStereoS16_2048Frames-12` | AMD Ryzen 5 5500U with Radeon Graphics | 2048 | 49716 | 114646 | 71.45
| `Nuked-OPL3` | `BenchmarkNukedOPL3GenerateStream_2048Frames` | AMD Ryzen 5 5500U with Radeon Graphics | 2048 | 49716 | 885166 | 9.25

Resampled (`44100 Hz`) result:

| Implementation | Benchmark | CPU | Iterations | Sample Rate | ns/op | MB/s |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `ImpSynth` | `BenchmarkGenerateStereoS16_2048Frames_44100Hz-12` | AMD Ryzen 5 5500U with Radeon Graphics | 2048 | 44100 | 136499 | 60.01
| `Nuked-OPL3` | `BenchmarkNukedOPL3GenerateStream_2048Frames` | AMD Ryzen 5 5500U with Radeon Graphics | 2048 | 44100 | 999243 | 8.20

At a glance:

- At `49716 Hz`, `ImpSynth` completed the same `2048`-frame render in about
  `7.72x` less time than `Nuked-OPL3` (`114646 ns/op` vs `885166 ns/op`).
- At `44100 Hz`, where both implementations exercise output resampling,
  `ImpSynth` completed the same render in about `7.32x` less time than
  `Nuked-OPL3` (`136499 ns/op` vs `999243 ns/op`).

Why `ImpSynth` is faster:

- `ImpSynth` is a Go-native synth tailored to the subset this repository uses:
  2-op voices, the register patterns driven by the DMX-style patches here, and
  direct stereo PCM generation.
- Its render loop iterates only active channels and drops silent voices from
  the hot path, instead of stepping the full chip model every sample.
- `Nuked-OPL3` is optimized for faithful chip emulation across the broader OPL3
  behavior surface, including timing behavior, write-buffer handling, stereo
  routing details, and other hardware quirks that `ImpSynth` does not attempt
  to reproduce in full.

Legend:

| Field | Meaning |
| --- | --- |
| `Implementation` | The synth backend being measured. |
| `Benchmark` | The benchmark identifier emitted by the tool used for that implementation. Go benchmarks include a `-N` suffix for the active `GOMAXPROCS` value. |
| `CPU` | Processor model reported by the Go benchmark runner. |
| `Iterations` | Fixed render operations used for this comparison. Both implementations were run for the same `2048` calls. |
| `ns/op` | Average nanoseconds per benchmark operation. Here, one operation is one `GenerateStereoS16(2048)` call. |
| `MB/s` | Effective throughput based on stereo 16-bit PCM bytes produced per operation. |

Hardware / software used for the measurement:

- CPU: AMD Ryzen 5 5500U with Radeon Graphics
- OS: Linux 6.8.0-101-generic (Ubuntu)
- Go: go1.25.0 linux/amd64

## License

LGPL-2.1

See [`LICENSE`](LICENSE). This repository includes work based in part on
`Nuked-OPL3`.
