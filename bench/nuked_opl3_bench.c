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

static void apply_benchmark_patch(opl3_chip *chip) {
    OPL3_WriteReg(chip, 0x01, 0x20);
    OPL3_WriteReg(chip, 0x20, 0x01);
    OPL3_WriteReg(chip, 0x23, 0x01);
    OPL3_WriteReg(chip, 0x40, 0x18);
    OPL3_WriteReg(chip, 0x43, 0x00);
    OPL3_WriteReg(chip, 0x60, 0xF4);
    OPL3_WriteReg(chip, 0x63, 0xF6);
    OPL3_WriteReg(chip, 0x80, 0x55);
    OPL3_WriteReg(chip, 0x83, 0x14);
    OPL3_WriteReg(chip, 0xC0, 0x30);
    OPL3_WriteReg(chip, 0xA0, 0x98);
    OPL3_WriteReg(chip, 0xB0, 0x31);
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

int main(int argc, char **argv) {
    static int16_t pcm[FRAMES_PER_OP * CHANNELS];
    opl3_chip chip;
    uint64_t iterations = DEFAULT_ITERATIONS;
    uint32_t sample_rate = DEFAULT_SAMPLE_RATE;
    uint64_t elapsed_ns;
    double ns_per_op;
    double mb_per_sec;

    if (argc > 3) {
        fprintf(stderr, "usage: %s [iterations] [sample_rate]\n", argv[0]);
        return 1;
    }
    if (argc >= 2) {
        iterations = parse_iterations(argv[1]);
    }
    if (argc == 3) {
        sample_rate = parse_sample_rate(argv[2]);
    }

    OPL3_Reset(&chip, sample_rate);
    apply_benchmark_patch(&chip);

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
    printf("ns/op\t%.0f\n", ns_per_op);
    printf("MB/s\t%.2f\n", mb_per_sec);
    return 0;
}
