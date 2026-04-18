# Shareware Channel Comparison

This document compares per-channel `ImpSynth` output against `Nuked-OPL3` using the exported Wolf3D and DOOM shareware sequence fixtures.

Method:
- Render a short high-energy window from each song.
- Isolate each OPL channel by filtering its channel and operator register writes plus shared chip-global registers.
- Compare `ImpSynth` and `Nuked-OPL3` with spectral cosine similarity, energy ratio, and max sample delta.

## Wolf3D Shareware

Source: `wolf3d-shareware.zip`
Tick rate: `700`
Channels analyzed: `9`

Worst observed channel: `SUSPENSE` channel `1` with spectral similarity `0.000`, energy ratio `0.000x`, max delta `1`.

| Song | Window Start | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SUSPENSE | 1024 | 7 | 0.580 | 1 | 0.000 | 0.000x |
| POW | 6144 | 8 | 0.994 | 6 | 0.962 | 1.029x |

### SUSPENSE

Window start: `1024` frames. Active channels: `7`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.000 | 0.000x | 1 |
| 4 | 0.000 | 0.000x | 1 |
| 3 | 0.136 | 0.963x | 1529 |
| 2 | 0.924 | 1.553x | 494 |
| 5 | 0.999 | 1.041x | 184 |
| 7 | 1.000 | 1.000x | 63 |

### POW

Window start: `6144` frames. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.962 | 1.029x | 1816 |
| 2 | 0.990 | 1.024x | 492 |
| 5 | 0.998 | 0.993x | 148 |
| 4 | 1.000 | 1.000x | 30 |
| 1 | 1.000 | 1.000x | 100 |
| 7 | 1.000 | 0.966x | 13 |

## 

Source: ``
Tick rate: `0`
Channels analyzed: `0`

No songs analyzed.

