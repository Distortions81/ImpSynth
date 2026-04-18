#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include "opl3.h"

typedef struct {
    uint16_t reg;
    uint8_t value;
    uint16_t delay;
} seq_event;

static uint32_t parse_u32(const char *value) {
    char *end = NULL;
    unsigned long parsed = strtoul(value, &end, 0);
    if (value == NULL || *value == '\0' || end == value || *end != '\0') {
        fprintf(stderr, "invalid integer: %s\n", value);
        exit(1);
    }
    return (uint32_t)parsed;
}

static seq_event *load_events(const char *path, size_t *count_out) {
    FILE *f = fopen(path, "rb");
    seq_event *events;
    long size;
    size_t count;
    if (!f) {
        fprintf(stderr, "open event file failed: %s\n", path);
        exit(1);
    }
    if (fseek(f, 0, SEEK_END) != 0) {
        fclose(f);
        fprintf(stderr, "seek event file failed\n");
        exit(1);
    }
    size = ftell(f);
    if (size < 0 || (size % 5) != 0) {
        fclose(f);
        fprintf(stderr, "event file size invalid: %ld\n", size);
        exit(1);
    }
    rewind(f);
    count = (size_t)(size / 5);
    events = (seq_event *)calloc(count ? count : 1, sizeof(seq_event));
    if (!events) {
        fclose(f);
        fprintf(stderr, "alloc events failed\n");
        exit(1);
    }
    for (size_t i = 0; i < count; i++) {
        uint8_t raw[5];
        if (fread(raw, 1, 5, f) != 5) {
            fclose(f);
            free(events);
            fprintf(stderr, "read event file failed\n");
            exit(1);
        }
        events[i].reg = (uint16_t)(raw[0] | (raw[1] << 8));
        events[i].value = raw[2];
        events[i].delay = (uint16_t)(raw[3] | (raw[4] << 8));
    }
    fclose(f);
    *count_out = count;
    return events;
}

int main(int argc, char **argv) {
    opl3_chip chip;
    uint32_t sample_rate;
    uint32_t tick_rate;
    uint32_t frames;
    size_t event_count = 0;
    seq_event *events;
    size_t event_index = 0;
    uint32_t frames_until_next = 0;

    if (argc != 5) {
        fprintf(stderr, "usage: %s <sample_rate> <tick_rate> <frames> <event_file>\n", argv[0]);
        return 1;
    }

    sample_rate = parse_u32(argv[1]);
    tick_rate = parse_u32(argv[2]);
    frames = parse_u32(argv[3]);
    if (tick_rate == 0) {
        fprintf(stderr, "tick_rate must be > 0\n");
        return 1;
    }
    events = load_events(argv[4], &event_count);
    if (event_count == 0) {
        free(events);
        return 0;
    }

    OPL3_Reset(&chip, sample_rate);

    for (uint32_t frame = 0; frame < frames; frame++) {
        int16_t pcm[2];
        while (frames_until_next == 0) {
            seq_event ev = events[event_index++];
            if (event_index >= event_count) {
                event_index = 0;
            }
            OPL3_WriteReg(&chip, ev.reg, ev.value);
            frames_until_next = (uint32_t)ev.delay * (sample_rate / tick_rate);
            if (frames_until_next > 0) {
                break;
            }
        }
        OPL3_Generate(&chip, pcm);
        printf("%d %d\n", pcm[0], pcm[1]);
        if (frames_until_next > 0) {
            frames_until_next--;
        }
    }

    free(events);
    return 0;
}
