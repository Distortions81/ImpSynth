# Shareware Attack Plan

This plan uses the current whole-song findings in [shareware-channel-findings.md](/home/dist/github/ImpSynth/docs/shareware-channel-findings.md:1) and orders work from worst mismatch to best.

## 1. Wolf3D melodic channels with very low spectral match

Highest-value targets:
- `SUSPENSE` ch `7` spec `0.052`
- `SUSPENSE` ch `4` spec `0.120`
- `NAZI_NOR` ch `1` spec `0.131`
- `URAHERO` ch `6` spec `0.135`
- `02-untitled` ch `2` spec `0.185`
- `POW` ch `7` spec `0.187`

Why this is first:
- These are the lowest whole-song matches in the report.
- Most of them still have energy ratios near `1.0x`, so the bug is more likely `tone/shape` than simple loudness.
- This is the best place to find remaining operator, phase, feedback, or waveform mismatches.

Plan:
1. Dump register traces for these channels and classify the patch type:
   `feedback`, `waveform`, `algorithm`, `vibrato/tremolo`, `note-select`, `sustain/release`.
2. Compare the worst one channel-by-channel against Nuked over time, not just aggregate score.
3. Prioritize channels that stay bad for most of the song over channels that only have a few bad sections.
4. Build one focused regression fixture per bug class once identified.

Expected likely causes:
- high-feedback behavior
- alternate waveform handling
- channel algorithm / modulator-carrier interaction
- envelope progression or retrigger edge cases

## 2. Wolf3D channels with large level mismatch

Highest-value targets:
- `SEARCHN` ch `1` energy `1.536x`
- `GETTHEM` ch `1` energy `1.530x`
- `GETTHEM` ch `2` energy `1.603x`
- `SUSPENSE` ch `2` energy `1.633x`
- `SUSPENSE` ch `3` energy `1.250x`
- `ENDLEVEL` ch `2` energy `0.559x`

Why this is next:
- These are likely easier to fix than the very-low-spectrum channels.
- We already know `GETTHEM` ch `1` is a long-running high-feedback case.
- The spectral match is often moderate here, which suggests the structure is mostly right and the remaining issue is gain/depth.

Plan:
1. Reuse the earlier `GETTHEM` high-feedback investigation as the first checkpoint.
2. Compare operator output ranges against Nuked for the same patch and note range.
3. Check whether the mismatch tracks:
   feedback amount, carrier modulation depth, or envelope attenuation.
4. Tune only after identifying which stage is hot or weak.

Expected likely causes:
- feedback scaling under OPL2-style voices
- operator output range
- envelope attenuation differences
- carrier modulation depth

## 3. Wolf3D channels with huge max-delta spikes

Highest-value targets:
- `02-untitled` ch `2` max delta `8159`
- `02-untitled` ch `3` max delta `8159`
- `02-untitled` ch `5` max delta `8159`
- `02-untitled` ch `7` max delta `7432`
- `SEARCHN` ch `1` max delta `4069`
- `GETTHEM` ch `1` max delta `2878`

Why this matters:
- These often indicate clipping, phase inversion regions, or abrupt state differences.
- They are useful for finding “one bad rule” rather than a generally fuzzy approximation.

Plan:
1. Isolate one chunk where the spike happens.
2. Inspect whether the spike is:
   sustained, transient, or tied to note-on/retrigger.
3. If it is transient, look at key-on ordering and envelope reset behavior.
4. If it is sustained, treat it as a patch math issue and fold it into section 1 or 2.

## 4. DOOM short intro outliers

Highest-value targets:
- `D_INTROA` ch `2` spec `0.253`
- `D_INTROA` ch `17` spec `0.275`
- `D_INTROA` ch `1` spec `0.331`
- `D_INTROA` ch `16` spec `0.359`
- `D_INTRO` ch `10` spec `0.605`

Why this is after Wolf:
- DOOM overall is already much closer than the worst Wolf songs.
- These intro tracks are short and should be easy to isolate.
- Their energy ratios are mostly sane, so this is again likely tone/path behavior rather than gross timing failure.

Plan:
1. Compare `D_INTROA` and `D_INTRO` patches against the Wolf failure classes.
2. Check whether the same bug class explains both Wolf and DOOM outliers.
3. Only create DOOM-specific fixes if they do not collapse into the same root cause.

## 5. DOOM broad “good but not perfect” long-song quality

Representative tracks:
- `D_INTER` mean spec `0.707`
- `D_E1M1` mean spec `0.791`
- `D_E1M9` mean spec `0.807`
- `D_E1M4` mean spec `0.839`

Why this is later:
- These tracks are not catastrophically wrong.
- They likely reflect the aggregate effect of smaller per-voice mismatches already visible in higher-priority items.
- Fixing the worst Wolf and intro cases may lift these automatically.

Plan:
1. Re-run the whole-song report after each major fix from sections 1-4.
2. Only drill into these songs if they remain stubborn after the targeted fixes land.

## 6. Near-match cleanup

Examples:
- `WONDERIN`
- `CORNER`
- `23-untitled`
- `D_VICTOR`
- `D_E1M5`

Why last:
- These already look good enough for practical compatibility.
- Time spent here is likely polishing, not unlocking bigger wins.

Plan:
1. Leave these alone until the worse cases are improved.
2. Use them as regression checks to ensure fixes do not cause collateral damage.

## Recommended work order

1. `SUSPENSE` ch `7`
2. `NAZI_NOR` ch `1`
3. `URAHERO` ch `6`
4. `02-untitled` ch `2`
5. `GETTHEM` ch `1`
6. `SEARCHN` ch `1`
7. `D_INTROA` ch `2`
8. `D_INTRO` ch `10`

## Success criteria

- Raise the worst Wolf channels above roughly `0.5` spectral similarity.
- Bring major Wolf energy outliers closer to `1.0x`, especially `GETTHEM` and `SEARCHN`.
- Raise `D_INTROA` out of the `0.25-0.35` range.
- Re-run the full whole-song report after each fix batch and reorder this plan based on the new worst channels.
