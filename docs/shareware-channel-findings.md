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

Worst observed channel: `SUSPENSE` channel `1` with spectral similarity `0.713`, energy ratio `0.992x`, max delta `241`.

| Song | Frames | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SUSPENSE | 5709891 | 7 | 0.839 | 1 | 0.713 | 0.992x |
| GETTHEM | 3977846 | 8 | 0.882 | 1 | 0.725 | 1.530x |
| 02-untitled | 3392167 | 8 | 0.895 | 1 | 0.725 | 1.551x |
| SEARCHN | 3265787 | 8 | 0.951 | 1 | 0.732 | 1.536x |
| URAHERO | 1004721 | 8 | 0.942 | 6 | 0.753 | 0.985x |
| POW | 3237103 | 8 | 0.937 | 7 | 0.753 | 0.951x |
| ENDLEVEL | 1000745 | 8 | 0.902 | 2 | 0.761 | 0.559x |
| NAZI_NOR | 2868400 | 6 | 0.925 | 1 | 0.764 | 0.986x |
| 23-untitled | 940608 | 6 | 0.928 | 6 | 0.788 | 0.994x |
| CORNER | 3045403 | 6 | 0.967 | 6 | 0.804 | 0.987x |
| WONDERIN | 3522239 | 8 | 0.996 | 1 | 0.988 | 0.999x |

### SUSPENSE

Full-song frames: `5709891`. Active channels: `7`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.713 | 0.992x | 241 |
| 2 | 0.753 | 1.633x | 782 |
| 7 | 0.791 | 0.957x | 354 |
| 4 | 0.806 | 1.015x | 2429 |
| 3 | 0.862 | 1.250x | 1180 |
| 6 | 0.961 | 0.993x | 1716 |

### GETTHEM

Full-song frames: `3977846`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.725 | 1.530x | 2878 |
| 2 | 0.753 | 1.603x | 545 |
| 3 | 0.838 | 1.000x | 2199 |
| 4 | 0.853 | 0.997x | 382 |
| 6 | 0.928 | 0.996x | 295 |
| 7 | 0.964 | 0.995x | 122 |

### 02-untitled

Full-song frames: `3392167`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.725 | 1.551x | 3737 |
| 2 | 0.785 | 1.049x | 8159 |
| 6 | 0.896 | 0.996x | 2888 |
| 3 | 0.908 | 1.047x | 8159 |
| 7 | 0.944 | 0.995x | 7432 |
| 4 | 0.962 | 1.010x | 7493 |

### SEARCHN

Full-song frames: `3265787`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.732 | 1.536x | 4069 |
| 2 | 0.924 | 0.997x | 204 |
| 6 | 0.978 | 0.996x | 850 |
| 8 | 0.990 | 0.990x | 566 |
| 7 | 0.991 | 0.975x | 340 |
| 4 | 0.998 | 0.997x | 800 |

### URAHERO

Full-song frames: `1004721`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.753 | 0.985x | 85 |
| 7 | 0.796 | 0.995x | 1540 |
| 8 | 0.993 | 0.976x | 30 |
| 5 | 0.996 | 0.999x | 801 |
| 1 | 0.998 | 1.000x | 565 |
| 2 | 0.998 | 0.998x | 165 |

### POW

Full-song frames: `3237103`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 7 | 0.753 | 0.951x | 127 |
| 1 | 0.868 | 0.999x | 996 |
| 6 | 0.886 | 0.995x | 1872 |
| 8 | 0.995 | 0.996x | 1215 |
| 2 | 0.998 | 1.008x | 553 |
| 5 | 0.999 | 1.010x | 248 |

### ENDLEVEL

Full-song frames: `1000745`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 2 | 0.761 | 0.559x | 1215 |
| 1 | 0.769 | 1.007x | 1215 |
| 7 | 0.791 | 0.996x | 1015 |
| 6 | 0.941 | 0.991x | 443 |
| 5 | 0.983 | 1.003x | 950 |
| 4 | 0.984 | 0.976x | 1032 |

### NAZI_NOR

Full-song frames: `2868400`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.764 | 0.986x | 97 |
| 6 | 0.798 | 0.991x | 1678 |
| 4 | 0.993 | 1.000x | 72 |
| 2 | 0.998 | 0.998x | 167 |
| 3 | 0.999 | 1.002x | 503 |
| 5 | 0.999 | 0.998x | 177 |

### 23-untitled

Full-song frames: `940608`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.788 | 0.994x | 1106 |
| 2 | 0.861 | 0.809x | 1873 |
| 5 | 0.959 | 0.999x | 121 |
| 3 | 0.974 | 1.002x | 117 |
| 4 | 0.992 | 1.000x | 565 |
| 1 | 0.998 | 1.001x | 704 |

### CORNER

Full-song frames: `3045403`. Active channels: `6`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.804 | 0.987x | 557 |
| 4 | 0.999 | 0.997x | 261 |
| 3 | 0.999 | 0.998x | 393 |
| 1 | 0.999 | 1.027x | 270 |
| 5 | 0.999 | 1.001x | 640 |
| 2 | 1.000 | 1.002x | 475 |

### WONDERIN

Full-song frames: `3522239`. Active channels: `8`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 1 | 0.988 | 0.999x | 990 |
| 6 | 0.992 | 0.991x | 1843 |
| 3 | 0.997 | 0.996x | 126 |
| 2 | 0.997 | 0.994x | 245 |
| 4 | 0.997 | 0.994x | 163 |
| 8 | 0.999 | 0.976x | 74 |

## DOOM Shareware

Source: `DOOM1.WAD`
Tick rate: `140`
Channels analyzed: `18`

Worst observed channel: `D_INTRO` channel `14` with spectral similarity `0.826`, energy ratio `0.998x`, max delta `3860`.

| Song | Frames | Active Channels | Mean Spec | Worst Channel | Worst Spec | Worst Energy Ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| D_INTRO | 340800 | 18 | 0.958 | 14 | 0.826 | 0.998x |
| D_INTROA | 340800 | 18 | 0.962 | 16 | 0.840 | 1.091x |
| D_E1M1 | 4771200 | 18 | 0.925 | 6 | 0.895 | 0.998x |
| D_INTER | 9930770 | 18 | 0.943 | 17 | 0.933 | 1.001x |
| D_E1M4 | 8482015 | 18 | 0.947 | 12 | 0.934 | 0.998x |
| D_E1M9 | 6827005 | 18 | 0.948 | 0 | 0.940 | 1.002x |
| D_E1M6 | 4174800 | 18 | 0.958 | 13 | 0.950 | 0.998x |
| D_E1M3 | 13518400 | 18 | 0.967 | 0 | 0.959 | 0.999x |
| D_E1M5 | 8150800 | 18 | 0.970 | 15 | 0.960 | 1.009x |
| D_VICTOR | 9542400 | 18 | 0.971 | 15 | 0.964 | 1.001x |
| D_E1M8 | 7554400 | 18 | 0.983 | 3 | 0.970 | 0.998x |
| D_E1M2 | 7715215 | 18 | 0.983 | 4 | 0.972 | 1.002x |
| D_E1M7 | 7497600 | 18 | 0.990 | 9 | 0.987 | 0.996x |

### D_INTRO

Full-song frames: `340800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 14 | 0.826 | 0.998x | 3860 |
| 0 | 0.834 | 1.001x | 3173 |
| 10 | 0.856 | 0.996x | 1337 |
| 12 | 0.926 | 1.000x | 2941 |
| 16 | 0.932 | 0.999x | 2429 |
| 15 | 0.942 | 1.000x | 2428 |

### D_INTROA

Full-song frames: `340800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 16 | 0.840 | 1.091x | 3860 |
| 1 | 0.844 | 1.087x | 4629 |
| 11 | 0.907 | 0.996x | 2428 |
| 12 | 0.909 | 0.999x | 2429 |
| 2 | 0.929 | 0.810x | 252 |
| 17 | 0.933 | 0.799x | 266 |

### D_E1M1

Full-song frames: `4771200`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 6 | 0.895 | 0.998x | 4397 |
| 5 | 0.902 | 1.010x | 4343 |
| 3 | 0.909 | 1.008x | 4357 |
| 4 | 0.911 | 1.000x | 4121 |
| 17 | 0.919 | 1.009x | 4451 |
| 2 | 0.924 | 1.006x | 4415 |

### D_INTER

Full-song frames: `9930770`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 17 | 0.933 | 1.001x | 4801 |
| 2 | 0.937 | 1.002x | 4735 |
| 1 | 0.939 | 1.003x | 4697 |
| 0 | 0.940 | 1.002x | 4747 |
| 5 | 0.940 | 1.002x | 4845 |
| 8 | 0.940 | 0.998x | 4735 |

### D_E1M4

Full-song frames: `8482015`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 12 | 0.934 | 0.998x | 4653 |
| 0 | 0.936 | 1.003x | 4735 |
| 17 | 0.939 | 1.003x | 4799 |
| 16 | 0.940 | 1.005x | 4653 |
| 2 | 0.943 | 1.001x | 4659 |
| 3 | 0.944 | 1.004x | 4857 |

### D_E1M9

Full-song frames: `6827005`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 0 | 0.940 | 1.002x | 4741 |
| 4 | 0.941 | 1.003x | 4825 |
| 3 | 0.941 | 1.001x | 4553 |
| 5 | 0.942 | 1.003x | 4839 |
| 1 | 0.943 | 1.003x | 4831 |
| 8 | 0.944 | 0.999x | 4845 |

### D_E1M6

Full-song frames: `4174800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 13 | 0.950 | 0.998x | 2682 |
| 8 | 0.952 | 0.998x | 4064 |
| 3 | 0.952 | 1.002x | 2476 |
| 14 | 0.952 | 0.997x | 3531 |
| 7 | 0.955 | 0.997x | 2487 |
| 5 | 0.956 | 1.002x | 3404 |

### D_E1M3

Full-song frames: `13518400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 0 | 0.959 | 0.999x | 4138 |
| 15 | 0.961 | 1.000x | 4010 |
| 1 | 0.962 | 1.000x | 3945 |
| 5 | 0.962 | 0.999x | 3172 |
| 12 | 0.963 | 0.997x | 3688 |
| 17 | 0.965 | 0.999x | 3290 |

### D_E1M5

Full-song frames: `8150800`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 15 | 0.960 | 1.009x | 1699 |
| 3 | 0.963 | 0.996x | 999 |
| 12 | 0.966 | 0.996x | 510 |
| 4 | 0.966 | 0.997x | 163 |
| 14 | 0.966 | 0.995x | 510 |
| 7 | 0.967 | 0.995x | 160 |

### D_VICTOR

Full-song frames: `9542400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 15 | 0.964 | 1.001x | 5227 |
| 0 | 0.964 | 1.002x | 4047 |
| 16 | 0.965 | 1.001x | 5263 |
| 5 | 0.966 | 1.001x | 3737 |
| 17 | 0.967 | 1.002x | 4813 |
| 2 | 0.967 | 1.000x | 3399 |

### D_E1M8

Full-song frames: `7554400`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 3 | 0.970 | 0.998x | 4793 |
| 5 | 0.971 | 0.996x | 4851 |
| 2 | 0.971 | 0.998x | 4801 |
| 15 | 0.971 | 0.994x | 4705 |
| 9 | 0.975 | 0.993x | 4851 |
| 14 | 0.980 | 0.995x | 4653 |

### D_E1M2

Full-song frames: `7715215`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 4 | 0.972 | 1.002x | 3211 |
| 17 | 0.974 | 1.003x | 3220 |
| 5 | 0.975 | 1.002x | 3063 |
| 15 | 0.980 | 1.001x | 3009 |
| 16 | 0.981 | 1.002x | 3286 |
| 6 | 0.982 | 0.998x | 3345 |

### D_E1M7

Full-song frames: `7497600`. Active channels: `18`. Worst channels first.

| Channel | Spec | Energy Ratio | Max Delta |
| --- | ---: | ---: | ---: |
| 9 | 0.987 | 0.996x | 4098 |
| 10 | 0.988 | 0.994x | 4309 |
| 13 | 0.988 | 0.999x | 4653 |
| 16 | 0.989 | 1.000x | 4761 |
| 4 | 0.989 | 1.000x | 4851 |
| 8 | 0.989 | 0.998x | 4274 |

