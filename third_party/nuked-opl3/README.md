Vendored `Nuked-OPL3` sources used by the benchmark harness.

Origin:
- Repository: `https://github.com/nukeykt/Nuked-OPL3`
- Commit: `cfedb09efc03f1d7b5fc1f04dd449d77d8c49d50`
- Files: `opl3.h`, `opl3.c`

These files are checked in so the comparison benchmarks remain runnable from a
clean checkout without relying on a live network fetch.

Refresh:
- Run `./scripts/fetch-nuked-opl3.sh` to re-fetch the same pinned revision into
  this directory.
