#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include "opl3.h"

static uint32_t parse_u32(const char *value) {
    char *end = NULL;
    unsigned long parsed = strtoul(value, &end, 0);
    if (value == NULL || *value == '\0' || end == value || *end != '\0') {
        fprintf(stderr, "invalid integer: %s\n", value);
        exit(1);
    }
    return (uint32_t)parsed;
}

static uint16_t parse_u16(const char *value) {
    char *end = NULL;
    unsigned long parsed = strtoul(value, &end, 0);
    if (value == NULL || *value == '\0' || end == value || *end != '\0' || parsed > 0x1ff) {
        fprintf(stderr, "invalid register: %s\n", value);
        exit(1);
    }
    return (uint16_t)parsed;
}

static uint8_t parse_u8(const char *value) {
    char *end = NULL;
    unsigned long parsed = strtoul(value, &end, 0);
    if (value == NULL || *value == '\0' || end == value || *end != '\0' || parsed > 0xff) {
        fprintf(stderr, "invalid value: %s\n", value);
        exit(1);
    }
    return (uint8_t)parsed;
}

int main(int argc, char **argv) {
    opl3_chip chip;
    uint32_t sample_rate;
    uint32_t frames;

    if (argc < 3 || ((argc - 3) % 2) != 0) {
        fprintf(stderr, "usage: %s <sample_rate> <frames> [reg value]...\n", argv[0]);
        return 1;
    }

    sample_rate = parse_u32(argv[1]);
    frames = parse_u32(argv[2]);

    OPL3_Reset(&chip, sample_rate);
    for (int i = 3; i + 1 < argc; i += 2) {
        OPL3_WriteReg(&chip, parse_u16(argv[i]), parse_u8(argv[i + 1]));
    }

    for (uint32_t i = 0; i < frames; i++) {
        int16_t pcm[2];
        OPL3_Generate(&chip, pcm);
        printf("%d %d\n", pcm[0], pcm[1]);
    }
    return 0;
}
