# Agent notes

## Creating a new song

Never copy an existing song folder. Start from the command:

    go run ./cmd/tracker new -name "my song" -channels 4 -rows 64 tracks/my-song

This writes `tracks/my-song/song.txt` with one empty pattern and no instruments. Then edit `song.txt`.

Build with `go run ./cmd/tracker build tracks/my-song`, which writes `tracks/my-song/song.xm`.

## Adding a pattern

1. Add a `Pattern N` block after the last pattern, with `Rows:` and one `Row NNN:` line per row. Each row has one cell per channel, separated by ` | `.
2. Increase `NumPatterns` by one.
3. Add the pattern number to `Order:`.

## Adding an instrument

1. Add an `Instrument N` block after the last instrument, numbered from 0. Copy an existing block's layout, and set `Keymap:` to 96 sample numbers, one per note.
2. Increase `NumInstruments` by one.
3. Under the instrument, add a `Sample N` block for each sample, numbered from 0, and set `Samples:` to the count.
4. Set `File:` to the sample path **without an extension**, e.g. `samples/01-00-lead`.
5. Create that sample as **one** of:
   - `samples/01-00-lead.txt`, a recipe (see `recipe/recipe.go` for the syntax), or
   - `samples/01-00-lead.wav`, mono 8-bit or 16-bit PCM.

   Never have both. Build fails if both exist.

To listen to a recipe, render it to a WAV outside the song folder:

    go run ./cmd/tracker render -o /tmp/lead.wav tracks/my-song/samples/01-00-lead.txt

## Rules

- Keep `NumPatterns`, `NumInstruments`, `Rows:` and `Samples:` in step with the blocks they count.
- Names are at most 22 characters for instruments and samples, and 20 for the song.
- Loop points in `song.txt` are in frames, not bytes.
- Stereo samples are not supported.

## Writing a song as Go code

A song can be written as a Go program instead of `song.txt`. Use the `song` package (`song/`) and put the program in the song folder as `song.go`, with `package main`. See `tracks/track3-go/song.go` for a full example.

Build it with `go run ./cmd/tracker build tracks/my-song`, which runs the program. The program writes its own `song.xm`, so `-o` doesn't apply. Don't write a generator script outside the repo to produce `song.txt`.

## Samples and mixing

- A volume-column value (`song.Vol`) replaces the sample's own volume for that note. The instrument's `Volume` is only the default when the cell has no volume column, so raising it alone changes nothing on notes that set a volume.
- A looping sample must hold a whole number of cycles of every layer, or it clicks at the loop point. Use frequencies that are multiples of `recipe.Rate` / frames, e.g. 261.34375, 522.6875 and 784.03125 for a 256-frame sample.
- Two copies of a wave with `Phase(0.5)` cancel all odd harmonics, the fundamental included: a saw pad made that way sounds an octave up and hollow. Use harmonic multiples for thickness instead, or a small phase offset when a little beating is wanted.
- The kick is tuned by note: a 200 Hz sine played at `C-2` sounds at 50 Hz. Don't write the wanted frequency into the recipe if the note is transposed.
