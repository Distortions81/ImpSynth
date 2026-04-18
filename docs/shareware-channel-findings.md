# Shareware Channel Comparison

This document compares per-channel `ImpSynth` output against `Nuked-OPL3` using the exported Wolf3D and DOOM shareware sequence fixtures.

Method:
- Render one full song cycle for each active OPL voice/channel in `ImpSynth` and `Nuked-OPL3`.
- Split the rendered PCM into in-memory chunks for spectral comparison.
- Aggregate chunk similarities plus full-song energy ratio and max sample delta.

## Wolf3D Shareware

Source: `wolf3d-shareware.zip`
Tick rate: `700`
Channels analyzed: `9`

Worst observed channel: `SUSPENSE` channel `7` with spectral similarity `0.052`, energy ratio `0.957x`, max delta `354`.

| Song | Frames | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SUSPENSE | 5709891 | 7 | 0.558 | 7 | 0.052 | 0.957x |
| NAZI_NOR | 2868400 | 6 | 0.818 | 1 | 0.131 | 0.986x |
| URAHERO | 1004721 | 8 | 0.858 | 6 | 0.135 | 0.985x |
| 02-untitled | 3392167 | 8 | 0.453 | 2 | 0.185 | 1.049x |
| POW | 3237103 | 8 | 0.728 | 7 | 0.187 | 0.951x |
| ENDLEVEL | 1000745 | 8 | 0.738 | 6 | 0.326 | 0.991x |
| SEARCHN | 3265787 | 8 | 0.725 | 7 | 0.359 | 0.975x |
| GETTHEM | 3977846 | 8 | 0.779 | 4 | 0.468 | 0.997x |
| 23-untitled | 940608 | 6 | 0.890 | 6 | 0.788 | 0.994x |
| CORNER | 3045403 | 6 | 0.973 | 6 | 0.848 | 0.987x |
| WONDERIN | 3522239 | 8 | 0.980 | 1 | 0.876 | 0.999x |

### SUSPENSE

Full-song frames: `5709891`. Active channels: `7`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 7 | 0.052 | 0.957x | 354 |
| 4 | 0.120 | 1.015x | 2429 |
| 1 | 0.219 | 0.992x | 241 |
| 6 | 0.837 | 0.993x | 1716 |
| 3 | 0.862 | 1.250x | 1180 |
| 2 | 0.906 | 1.633x | 782 |

### NAZI_NOR

Full-song frames: `2868400`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.131 | 0.986x | 97 |
| 6 | 0.802 | 0.991x | 1678 |
| 4 | 0.979 | 1.000x | 72 |
| 2 | 0.998 | 0.998x | 167 |
| 3 | 0.999 | 1.002x | 503 |
| 5 | 0.999 | 0.998x | 177 |

### URAHERO

Full-song frames: `1004721`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.135 | 0.985x | 85 |
| 7 | 0.824 | 0.995x | 1540 |
| 8 | 0.926 | 0.976x | 30 |
| 5 | 0.983 | 0.999x | 801 |
| 2 | 0.997 | 0.998x | 165 |
| 3 | 0.997 | 0.998x | 178 |

### 02-untitled

Full-song frames: `3392167`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.185 | 1.049x | 8159 |
| 7 | 0.269 | 0.995x | 7432 |
| 3 | 0.369 | 1.047x | 8159 |
| 6 | 0.404 | 0.996x | 2888 |
| 5 | 0.476 | 1.011x | 8159 |
| 4 | 0.518 | 1.010x | 7493 |

### POW

Full-song frames: `3237103`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 7 | 0.187 | 0.951x | 127 |
| 6 | 0.413 | 0.995x | 1872 |
| 1 | 0.421 | 0.999x | 996 |
| 8 | 0.805 | 0.996x | 1215 |
| 2 | 0.997 | 1.008x | 553 |
| 5 | 0.999 | 1.010x | 248 |

### ENDLEVEL

Full-song frames: `1000745`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.326 | 0.991x | 443 |
| 5 | 0.520 | 1.003x | 950 |
| 2 | 0.625 | 0.559x | 1215 |
| 1 | 0.660 | 1.007x | 1215 |
| 7 | 0.796 | 0.996x | 1015 |
| 4 | 0.984 | 0.976x | 1032 |

### SEARCHN

Full-song frames: `3265787`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 7 | 0.359 | 0.975x | 340 |
| 2 | 0.394 | 0.997x | 204 |
| 1 | 0.740 | 1.536x | 4069 |
| 6 | 0.772 | 0.996x | 850 |
| 4 | 0.800 | 0.997x | 800 |
| 5 | 0.809 | 0.993x | 353 |

### GETTHEM

Full-song frames: `3977846`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 4 | 0.468 | 0.997x | 382 |
| 3 | 0.640 | 1.000x | 2199 |
| 6 | 0.661 | 0.996x | 295 |
| 1 | 0.729 | 1.530x | 2878 |
| 7 | 0.842 | 0.995x | 122 |
| 2 | 0.916 | 1.603x | 545 |

### 23-untitled

Full-song frames: `940608`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.788 | 0.994x | 1106 |
| 5 | 0.838 | 0.999x | 121 |
| 2 | 0.861 | 0.809x | 1873 |
| 4 | 0.915 | 1.000x | 565 |
| 3 | 0.942 | 1.002x | 117 |
| 1 | 0.998 | 1.001x | 704 |

### CORNER

Full-song frames: `3045403`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.848 | 0.987x | 557 |
| 3 | 0.995 | 0.998x | 393 |
| 4 | 0.999 | 0.997x | 261 |
| 1 | 0.999 | 1.027x | 270 |
| 5 | 0.999 | 1.001x | 640 |
| 2 | 1.000 | 1.002x | 475 |

### WONDERIN

Full-song frames: `3522239`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.876 | 0.999x | 990 |
| 6 | 0.992 | 0.991x | 1843 |
| 4 | 0.992 | 0.994x | 163 |
| 3 | 0.992 | 0.996x | 126 |
| 2 | 0.992 | 0.994x | 245 |
| 8 | 0.998 | 0.976x | 74 |

## DOOM Shareware

Source: `DOOM1.WAD`
Tick rate: `140`
Channels analyzed: `18`

Worst observed channel: `D_INTROA` channel `2` with spectral similarity `0.253`, energy ratio `0.810x`, max delta `252`.

| Song | Frames | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| D_INTROA | 340800 | 18 | 0.752 | 2 | 0.253 | 0.810x |
| D_INTRO | 340800 | 18 | 0.901 | 10 | 0.605 | 0.996x |
| D_INTER | 9930770 | 18 | 0.707 | 5 | 0.681 | 1.002x |
| D_E1M1 | 4771200 | 18 | 0.791 | 16 | 0.734 | 1.005x |
| D_E1M9 | 6827005 | 18 | 0.807 | 12 | 0.776 | 0.997x |
| D_E1M4 | 8482015 | 18 | 0.839 | 12 | 0.788 | 0.998x |
| D_E1M6 | 4174800 | 18 | 0.887 | 8 | 0.836 | 0.998x |
| D_E1M3 | 13518400 | 18 | 0.862 | 15 | 0.840 | 1.000x |
| D_E1M2 | 7715215 | 18 | 0.892 | 4 | 0.841 | 1.002x |
| D_E1M8 | 7554400 | 18 | 0.916 | 9 | 0.849 | 0.993x |
| D_E1M7 | 7497600 | 18 | 0.923 | 11 | 0.895 | 0.999x |
| D_VICTOR | 9542400 | 18 | 0.926 | 5 | 0.908 | 1.001x |
| D_E1M5 | 8150800 | 18 | 0.938 | 15 | 0.925 | 1.009x |

### D_INTROA

Full-song frames: `340800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.253 | 0.810x | 252 |
| 17 | 0.275 | 0.799x | 266 |
| 1 | 0.331 | 1.087x | 4629 |
| 16 | 0.359 | 1.091x | 3860 |
| 11 | 0.483 | 0.996x | 2428 |
| 15 | 0.485 | 0.986x | 1077 |

### D_INTRO

Full-song frames: `340800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 10 | 0.605 | 0.996x | 1337 |
| 17 | 0.737 | 1.013x | 2429 |
| 14 | 0.760 | 0.998x | 3860 |
| 16 | 0.819 | 0.999x | 2429 |
| 0 | 0.827 | 1.001x | 3173 |
| 1 | 0.872 | 1.001x | 2277 |

### D_INTER

Full-song frames: `9930770`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 5 | 0.681 | 1.002x | 4845 |
| 17 | 0.681 | 1.001x | 4801 |
| 7 | 0.682 | 0.998x | 4767 |
| 1 | 0.683 | 1.003x | 4697 |
| 4 | 0.691 | 1.003x | 4813 |
| 12 | 0.691 | 0.998x | 4775 |

### D_E1M1

Full-song frames: `4771200`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 16 | 0.734 | 1.005x | 4193 |
| 5 | 0.741 | 1.010x | 4343 |
| 2 | 0.742 | 1.006x | 4415 |
| 7 | 0.744 | 0.998x | 4355 |
| 3 | 0.746 | 1.008x | 4357 |
| 6 | 0.752 | 0.998x | 4397 |

### D_E1M9

Full-song frames: `6827005`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 12 | 0.776 | 0.997x | 4769 |
| 6 | 0.778 | 0.998x | 4735 |
| 8 | 0.781 | 0.999x | 4845 |
| 7 | 0.792 | 0.998x | 4709 |
| 10 | 0.795 | 0.998x | 4799 |
| 0 | 0.798 | 1.002x | 4741 |

### D_E1M4

Full-song frames: `8482015`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 12 | 0.788 | 0.998x | 4653 |
| 15 | 0.793 | 1.005x | 4703 |
| 7 | 0.803 | 0.997x | 4591 |
| 16 | 0.817 | 1.005x | 4653 |
| 10 | 0.821 | 0.997x | 4799 |
| 13 | 0.833 | 0.998x | 4845 |

### D_E1M6

Full-song frames: `4174800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 8 | 0.836 | 0.998x | 4064 |
| 4 | 0.862 | 1.002x | 2929 |
| 13 | 0.871 | 0.998x | 2682 |
| 11 | 0.871 | 0.998x | 2725 |
| 5 | 0.874 | 1.002x | 3404 |
| 1 | 0.882 | 1.001x | 3418 |

### D_E1M3

Full-song frames: `13518400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 15 | 0.840 | 1.000x | 4010 |
| 11 | 0.840 | 0.998x | 4006 |
| 12 | 0.846 | 0.997x | 3688 |
| 1 | 0.849 | 1.000x | 3945 |
| 7 | 0.853 | 0.998x | 3252 |
| 5 | 0.858 | 0.999x | 3172 |

### D_E1M2

Full-song frames: `7715215`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 4 | 0.841 | 1.002x | 3211 |
| 14 | 0.845 | 0.998x | 3194 |
| 5 | 0.855 | 1.002x | 3063 |
| 16 | 0.868 | 1.002x | 3286 |
| 9 | 0.870 | 0.998x | 3282 |
| 6 | 0.870 | 0.998x | 3345 |

### D_E1M8

Full-song frames: `7554400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 9 | 0.849 | 0.993x | 4851 |
| 4 | 0.862 | 1.001x | 4845 |
| 15 | 0.875 | 0.994x | 4705 |
| 12 | 0.896 | 0.999x | 4813 |
| 3 | 0.904 | 0.998x | 4793 |
| 5 | 0.905 | 0.996x | 4851 |

### D_E1M7

Full-song frames: `7497600`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 11 | 0.895 | 0.999x | 4599 |
| 2 | 0.900 | 1.001x | 4659 |
| 4 | 0.904 | 1.000x | 4851 |
| 10 | 0.908 | 0.994x | 4309 |
| 16 | 0.914 | 1.000x | 4761 |
| 14 | 0.914 | 0.998x | 4274 |

### D_VICTOR

Full-song frames: `9542400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 5 | 0.908 | 1.001x | 3737 |
| 11 | 0.911 | 0.997x | 3961 |
| 17 | 0.914 | 1.002x | 4813 |
| 15 | 0.917 | 1.001x | 5227 |
| 10 | 0.918 | 0.997x | 4349 |
| 0 | 0.920 | 1.002x | 4047 |

### D_E1M5

Full-song frames: `8150800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 15 | 0.925 | 1.009x | 1699 |
| 0 | 0.927 | 0.997x | 176 |
| 3 | 0.929 | 0.996x | 999 |
| 11 | 0.933 | 0.995x | 787 |
| 12 | 0.934 | 0.996x | 510 |
| 14 | 0.935 | 0.995x | 510 |

