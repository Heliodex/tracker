#!/usr/bin/env bash
# Builds each song, renders its song.xm with openmpt123 (apt package openmpt123), and encodes it to song.opus
# in the same folder with ffmpeg's libopus. The intermediate WAV is removed.
#
# Usage:
#   scripts/render-wav.sh              render every track in tracks/
#   scripts/render-wav.sh tracks/track10 tracks/track3
#
# The Opus is 48 kHz stereo, the same format as the existing song.opus files.
set -euo pipefail

cd "$(dirname "$0")/.."

for tool in openmpt123 ffmpeg; do
	if ! command -v "$tool" >/dev/null; then
		echo "$tool is not installed" >&2
		exit 1
	fi
done

if [ "$#" -eq 0 ]; then
	set -- tracks/*/
fi

for dir in "$@"; do
	dir=${dir%/}

	if [ -f "$dir/song.go" ] || [ -f "$dir/song.txt" ]; then
		go run ./cmd/tracker build -f "$dir"
	fi

	if [ ! -f "$dir/song.xm" ]; then
		echo "skipping $dir: no song.xm" >&2
		continue
	fi

	# openmpt123 names its output after the input, so render to song.xm.wav, encode it, then remove it
	openmpt123 --render --output-type wav --no-float --samplerate 48000 --force \
		--no-progress --no-meters --no-details "$dir/song.xm" >/dev/null

	ffmpeg -y -loglevel error -i "$dir/song.xm.wav" -c:a libopus "$dir/song.opus"
	rm -f "$dir/song.xm.wav"

	echo "$dir/song.opus"
done
