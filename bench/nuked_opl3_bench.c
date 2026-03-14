#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/time.h>

#include "opl3.h"

#define FRAMES_PER_OP 2048
#define CHANNELS 2
#define BYTES_PER_SAMPLE 2
#define DEFAULT_ITERATIONS 2048ULL
#define DEFAULT_SAMPLE_RATE 49716U
#define DEFAULT_VOICE_COUNT 3
#define MAX_VOICE_COUNT 18

static void apply_benchmark_voice(opl3_chip *chip, int ch) {
    static const uint16_t mod_slots[9] = {0, 1, 2, 8, 9, 10, 16, 17, 18};
    static const uint16_t car_slots[9] = {3, 4, 5, 11, 12, 13, 19, 20, 21};
    uint16_t base = 0;
    int local_ch = ch;
    uint16_t mod;
    uint16_t car;

    if (ch >= 9) {
        base = 0x100;
        local_ch = ch - 9;
    }

    mod = mod_slots[local_ch];
    car = car_slots[local_ch];

    OPL3_WriteReg(chip, base + 0x20 + mod, 0x01);
    OPL3_WriteReg(chip, base + 0x20 + car, 0x01);
    OPL3_WriteReg(chip, base + 0x40 + mod, 0x18);
    OPL3_WriteReg(chip, base + 0x40 + car, 0x00);
    OPL3_WriteReg(chip, base + 0x60 + mod, 0xF4);
    OPL3_WriteReg(chip, base + 0x60 + car, 0xF6);
    OPL3_WriteReg(chip, base + 0x80 + mod, 0x55);
    OPL3_WriteReg(chip, base + 0x80 + car, 0x14);
    OPL3_WriteReg(chip, base + 0xC0 + local_ch, 0x30);
    OPL3_WriteReg(chip, base + 0xA0 + local_ch, 0x98);
    OPL3_WriteReg(chip, base + 0xB0 + local_ch, 0x31);
}

static void apply_benchmark_patch(opl3_chip *chip, int voice_count) {
    OPL3_WriteReg(chip, 0x01, 0x20);
    for (int ch = 0; ch < voice_count; ch++) {
        apply_benchmark_voice(chip, ch);
    }
}

static uint64_t monotonic_ns(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return ((uint64_t)tv.tv_sec * 1000000000ull) + ((uint64_t)tv.tv_usec * 1000ull);
}

static uint64_t parse_iterations(const char *value) {
    char *end = NULL;
    unsigned long long parsed;

    if (value == NULL || *value == '\0') {
        return DEFAULT_ITERATIONS;
    }
    parsed = strtoull(value, &end, 10);
    if (end == value || *end != '\0' || parsed == 0) {
        fprintf(stderr, "invalid iteration count: %s\n", value);
        exit(1);
    }
    return (uint64_t)parsed;
}

static uint32_t parse_sample_rate(const char *value) {
    char *end = NULL;
    unsigned long parsed;

    if (value == NULL || *value == '\0') {
        return DEFAULT_SAMPLE_RATE;
    }
    parsed = strtoul(value, &end, 10);
    if (end == value || *end != '\0' || parsed == 0) {
        fprintf(stderr, "invalid sample rate: %s\n", value);
        exit(1);
    }
    return (uint32_t)parsed;
}

static int parse_voice_count(const char *value) {
    char *end = NULL;
    long parsed;

    if (value == NULL || *value == '\0') {
        return DEFAULT_VOICE_COUNT;
    }
    parsed = strtol(value, &end, 10);
    if (end == value || *end != '\0' || parsed < 1 || parsed > MAX_VOICE_COUNT) {
        fprintf(stderr, "invalid voice count: %s\n", value);
        exit(1);
    }
    return (int)parsed;
}

int main(int argc, char **argv) {
    static int16_t pcm[FRAMES_PER_OP * CHANNELS];
    opl3_chip chip;
    uint64_t iterations = DEFAULT_ITERATIONS;
    uint32_t sample_rate = DEFAULT_SAMPLE_RATE;
    int voice_count = DEFAULT_VOICE_COUNT;
    uint64_t elapsed_ns;
    double ns_per_op;
    double mb_per_sec;

    if (argc > 4) {
        fprintf(stderr, "usage: %s [iterations] [sample_rate] [voices]\n", argv[0]);
        return 1;
    }
    if (argc >= 2) {
        iterations = parse_iterations(argv[1]);
    }
    if (argc == 3) {
        sample_rate = parse_sample_rate(argv[2]);
    } else if (argc >= 4) {
        sample_rate = parse_sample_rate(argv[2]);
        voice_count = parse_voice_count(argv[3]);
    }

    OPL3_Reset(&chip, sample_rate);
    apply_benchmark_patch(&chip, voice_count);

    for (int i = 0; i < 2000; i++) {
        if (sample_rate == DEFAULT_SAMPLE_RATE) {
            for (int frame = 0; frame < FRAMES_PER_OP; frame++) {
                OPL3_Generate(&chip, &pcm[frame * CHANNELS]);
            }
        } else {
            OPL3_GenerateStream(&chip, pcm, FRAMES_PER_OP);
        }
    }

    elapsed_ns = monotonic_ns();
    for (uint64_t i = 0; i < iterations; i++) {
        if (sample_rate == DEFAULT_SAMPLE_RATE) {
            for (int frame = 0; frame < FRAMES_PER_OP; frame++) {
                OPL3_Generate(&chip, &pcm[frame * CHANNELS]);
            }
        } else {
            OPL3_GenerateStream(&chip, pcm, FRAMES_PER_OP);
        }
    }
    elapsed_ns = monotonic_ns() - elapsed_ns;

    ns_per_op = (double)elapsed_ns / (double)iterations;
    mb_per_sec = ((double)(FRAMES_PER_OP * CHANNELS * BYTES_PER_SAMPLE) * 1000.0) / ns_per_op;

    printf("BenchmarkNukedOPL3GenerateStream_2048Frames\n");
    printf("iterations\t%llu\n", (unsigned long long)iterations);
    printf("sample_rate\t%u\n", sample_rate);
    printf("voices\t%d\n", voice_count);
    printf("ns/op\t%.0f\n", ns_per_op);
    printf("MB/s\t%.2f\n", mb_per_sec);
    return 0;
}
