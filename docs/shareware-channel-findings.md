# Shareware Channel Comparison

This document compares per-channel `ImpSynth` output against `Nuked-OPL3` using the exported Wolf3D and DOOM shareware sequence fixtures.

Method:
- Pick one high-energy analysis window per song.
- Render each active OPL voice/channel once for `ImpSynth` and once for `Nuked-OPL3`.
- Compare the stereo PCM for that voice window using spectral cosine similarity, energy ratio, and max sample delta.

## Wolf3D Shareware

Source: `wolf3d-shareware.zip`
Tick rate: `700`
Channels analyzed: `9`

Worst observed channel: `` channel `1` with spectral similarity `0.000`, energy ratio `1.000x`, max delta `53`.

| Song | Window Start | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
|  | 1024 | 2 | 0.500 | 1 | 0.000 | 1.000x |
|  | 3072 | 3 | 0.667 | 6 | 0.000 | 1.000x |
| ENDLEVEL | 1024 | 3 | 0.666 | 7 | 0.000 | 0.998x |
| GETTHEM | 2048 | 2 | 0.452 | 4 | 0.000 | 0.999x |
| NAZI_NOR | 2048 | 6 | 0.662 | 5 | 0.000 | 0.999x |
| SEARCHN | 1024 | 2 | 0.454 | 2 | 0.000 | 1.000x |
| SUSPENSE | 1024 | 3 | 0.333 | 5 | 0.000 | 1.003x |
| WONDERIN | 1024 | 4 | 0.250 | 5 | 0.000 | 1.001x |
| URAHERO | 5120 | 8 | 1.000 | 8 | 0.997 | 0.973x |
| POW | 6144 | 5 | 1.000 | 2 | 0.998 | 1.016x |
| CORNER | 1024 | 1 | 1.000 | 1 | 1.000 | 1.005x |

### 

Window start: `1024` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.000 | 1.000x | 53 |
| 2 | 1.000 | 1.000x | 482 |

### 

Window start: `3072` frames. Active channels: `3`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.000 | 1.000x | 1106 |
| 1 | 1.000 | 1.002x | 338 |
| 2 | 1.000 | 1.000x | 63 |

### ENDLEVEL

Window start: `1024` frames. Active channels: `3`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 7 | 0.000 | 0.998x | 992 |
| 1 | 0.999 | 1.002x | 894 |
| 8 | 1.000 | 1.000x | 46 |

### GETTHEM

Window start: `2048` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 4 | 0.000 | 0.999x | 45 |
| 1 | 0.903 | 1.523x | 1232 |

### NAZI_NOR

Window start: `2048` frames. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 5 | 0.000 | 0.999x | 12 |
| 6 | 0.000 | 1.000x | 996 |
| 4 | 0.969 | 1.000x | 54 |
| 1 | 1.000 | 0.999x | 51 |
| 3 | 1.000 | 1.000x | 173 |
| 2 | 1.000 | 0.999x | 17 |

### SEARCHN

Window start: `1024` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.000 | 1.000x | 163 |
| 1 | 0.908 | 1.535x | 2838 |

### SUSPENSE

Window start: `1024` frames. Active channels: `3`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 5 | 0.000 | 1.003x | 51 |
| 7 | 0.000 | 1.002x | 301 |
| 1 | 1.000 | 1.000x | 210 |

### WONDERIN

Window start: `1024` frames. Active channels: `4`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 5 | 0.000 | 1.001x | 56 |
| 6 | 0.000 | 0.934x | 5 |
| 8 | 0.000 | 1.000x | 1 |
| 1 | 1.000 | 1.000x | 24 |

### URAHERO

Window start: `5120` frames. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 8 | 0.997 | 0.973x | 18 |
| 7 | 1.000 | 0.995x | 492 |
| 1 | 1.000 | 1.000x | 177 |
| 6 | 1.000 | 0.998x | 23 |
| 2 | 1.000 | 0.998x | 56 |
| 4 | 1.000 | 0.999x | 49 |

### POW

Window start: `6144` frames. Active channels: `5`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.998 | 1.016x | 377 |
| 4 | 1.000 | 1.001x | 35 |
| 3 | 1.000 | 1.001x | 25 |
| 1 | 1.000 | 1.000x | 117 |
| 6 | 1.000 | 0.998x | 13 |

### CORNER

Window start: `1024` frames. Active channels: `1`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 1.000 | 1.005x | 56 |

## DOOM Shareware

Source: `DOOM1.WAD`
Tick rate: `140`
Channels analyzed: `18`

Worst observed channel: `D_INTER` channel `3` with spectral similarity `0.000`, energy ratio `0.000x`, max delta `1`.

| Song | Window Start | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| D_INTER | 9216 | 14 | 0.713 | 3 | 0.000 | 0.000x |
| D_INTRO | 0 | 14 | 0.928 | 11 | 0.000 | 0.000x |
| D_INTROA | 13312 | 18 | 0.889 | 1 | 0.000 | 0.000x |
| D_E1M2 | 14336 | 4 | 0.870 | 3 | 0.555 | 0.549x |
| D_E1M1 | 0 | 6 | 0.984 | 2 | 0.904 | 1.160x |
| D_E1M4 | 0 | 5 | 0.993 | 3 | 0.966 | 0.998x |
| D_E1M8 | 0 | 3 | 0.988 | 1 | 0.968 | 0.944x |
| D_E1M9 | 3072 | 4 | 0.998 | 0 | 0.996 | 0.998x |
| D_E1M5 | 14336 | 4 | 0.998 | 0 | 0.996 | 0.660x |
| D_VICTOR | 14336 | 2 | 0.999 | 1 | 0.999 | 0.998x |
| D_E1M3 | 0 | 2 | 1.000 | 1 | 1.000 | 1.001x |
| D_E1M6 | 3072 | 8 | 1.000 | 3 | 1.000 | 0.999x |
| D_E1M7 | 13312 | 2 | 1.000 | 0 | 1.000 | 0.986x |

### D_INTER

Window start: `9216` frames. Active channels: `14`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 3 | 0.000 | 0.000x | 1 |
| 4 | 0.000 | 0.000x | 1 |
| 13 | 0.000 | 1.000x | 1 |
| 15 | 0.000 | 1.003x | 1837 |
| 5 | 0.981 | 0.649x | 2 |
| 8 | 0.994 | 0.999x | 2829 |

### D_INTRO

Window start: `0` frames. Active channels: `14`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 11 | 0.000 | 0.000x | 1 |
| 1 | 0.998 | 1.002x | 2277 |
| 7 | 1.000 | 1.000x | 2228 |
| 6 | 1.000 | 1.000x | 2228 |
| 17 | 1.000 | 1.102x | 2429 |
| 12 | 1.000 | 0.999x | 712 |

### D_INTROA

Window start: `13312` frames. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.000 | 0.000x | 1 |
| 2 | 0.000 | 0.000x | 1 |
| 0 | 1.000 | 0.988x | 7 |
| 15 | 1.000 | 0.997x | 18 |
| 7 | 1.000 | 0.999x | 2042 |
| 17 | 1.000 | 0.444x | 1 |

### D_E1M2

Window start: `14336` frames. Active channels: `4`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 3 | 0.555 | 0.549x | 4 |
| 2 | 0.923 | 0.506x | 4 |
| 0 | 1.000 | 0.999x | 1833 |
| 1 | 1.000 | 0.999x | 2042 |

### D_E1M1

Window start: `0` frames. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.904 | 1.160x | 3569 |
| 3 | 0.999 | 1.003x | 1 |
| 1 | 1.000 | 1.001x | 999 |
| 0 | 1.000 | 1.001x | 191 |
| 4 | 1.000 | 1.000x | 74 |
| 5 | 1.000 | 0.999x | 1717 |

### D_E1M4

Window start: `0` frames. Active channels: `5`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 3 | 0.966 | 0.998x | 893 |
| 1 | 0.999 | 0.999x | 1110 |
| 2 | 1.000 | 1.000x | 653 |
| 0 | 1.000 | 1.001x | 125 |
| 4 | 1.000 | 0.998x | 1114 |

### D_E1M8

Window start: `0` frames. Active channels: `3`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.968 | 0.944x | 3 |
| 2 | 0.995 | 0.861x | 3 |
| 0 | 1.000 | 1.000x | 1576 |

### D_E1M9

Window start: `3072` frames. Active channels: `4`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 0 | 0.996 | 0.998x | 1440 |
| 1 | 0.998 | 0.999x | 1114 |
| 3 | 1.000 | 0.997x | 1114 |
| 2 | 1.000 | 1.000x | 101 |

### D_E1M5

Window start: `14336` frames. Active channels: `4`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 0 | 0.996 | 0.660x | 7 |
| 1 | 0.997 | 0.656x | 7 |
| 2 | 1.000 | 0.986x | 124 |
| 3 | 1.000 | 0.987x | 124 |

### D_VICTOR

Window start: `14336` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.999 | 0.998x | 368 |
| 0 | 1.000 | 0.997x | 245 |

### D_E1M3

Window start: `0` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 1.000 | 1.001x | 545 |
| 0 | 1.000 | 1.002x | 608 |

### D_E1M6

Window start: `3072` frames. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 3 | 1.000 | 0.999x | 962 |
| 5 | 1.000 | 0.999x | 689 |
| 1 | 1.000 | 0.998x | 702 |
| 6 | 1.000 | 1.000x | 31 |
| 2 | 1.000 | 0.998x | 284 |
| 4 | 1.000 | 0.998x | 175 |

### D_E1M7

Window start: `13312` frames. Active channels: `2`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 0 | 1.000 | 0.986x | 11 |
| 1 | 1.000 | 0.986x | 11 |

