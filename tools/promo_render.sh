#!/usr/bin/env bash
# Docker內FFmpeg合成；固定63秒，來源與分鏡由promo_assets.py保存。
set -euo pipefail
test -f /.dockerenv
version=${1:-v.0.1.2-20261102}
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
video="/out/pod-${version}-promo.mp4"
test -f /out/assets/s00.png
test -f /capture/intro/audio.s16le
test ! -e /out/segments
mkdir /out/segments /out/review
F=(ffmpeg -nostdin -y -hide_banner -loglevel warning -threads 1 -filter_threads 1)
V=(-c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p -r 25 -threads 1)
still(){
 "${F[@]}" -loop 1 -framerate 25 -i "/out/assets/$1" -t "$2" \
  -vf "fade=t=in:st=0:d=0.2,fade=t=out:st=$3:d=0.2,format=yuv420p" "${V[@]}" "/out/segments/$4.mp4"
}
still s00.png 6 5.8 s00
still s01.png 8 7.8 s01
still s02.png 7 6.8 s02
for part in map guide;do still "tools-$part.png" 4 3.8 "tools-$part";done
printf "file '/out/segments/tools-map.mp4'\nfile '/out/segments/tools-guide.mp4'\n" > /out/segments/tools.txt
"${F[@]}" -f concat -safe 0 -i /out/segments/tools.txt -c copy /out/segments/s03.mp4
"${F[@]}" -framerate 25 -i /out/assets/switch/%04d.png -frames:v 200 "${V[@]}" /out/segments/s04.mp4
for i in 0 1 2 3 4;do still "lang$i.png" 1.4 1.2 "lang$i";done
printf "file '/out/segments/lang%s.mp4'\n" 0 1 2 3 4 > /out/segments/langs.txt
"${F[@]}" -f concat -safe 0 -i /out/segments/langs.txt -c copy /out/segments/s05.mp4
still s06.png 7 6.8 s06
for i in 0 1 2;do still "ending$i.png" 2 1.8 "ending$i";done
printf "file '/out/segments/ending%s.mp4'\n" 0 1 2 > /out/segments/endings.txt
"${F[@]}" -f concat -safe 0 -i /out/segments/endings.txt -c copy /out/segments/s07.mp4
still s08.png 6 5.8 s08
printf "file '/out/segments/s%02d.mp4'\n" 0 1 2 3 4 5 6 7 8 > /out/segments/list.txt
"${F[@]}" -f concat -safe 0 -i /out/segments/list.txt -c copy /out/silent.mp4
"${F[@]}" -i /out/silent.mp4 -f s16le -ar 48000 -ac 2 -i /capture/intro/audio.s16le \
 -map 0:v -map 1:a -t 63 -c:v copy -af 'volume=2,atrim=end=63,afade=t=in:st=0:d=2,afade=t=out:st=60:d=3' \
 -c:a aac -b:a 192k -ar 48000 -ac 2 -movflags +faststart "$video"
ffprobe -v error -show_format -show_streams -of json "$video" > /out/ffprobe.json
ffmpeg -nostdin -hide_banner -threads 1 -i "$video" \
 -vn -af 'silencedetect=noise=-50dB:d=0.4,volumedetect' -f null - > /out/audio-check.log 2>&1
ffmpeg -nostdin -hide_banner -threads 1 -filter_threads 1 -i "$video" \
 -an -vf 'blackdetect=d=0.2:pix_th=0.02:pic_th=0.98,freezedetect=n=-60dB:d=2' -f null - > /out/image-check.log 2>&1
i=0
for t in 3 10 17.5 27 31 40.5 47.5 56 60;do
 "${F[@]}" -ss "$t" -i "$video" -frames:v 1 "/out/review/f$i.png"
 i=$((i+1))
done
"${F[@]}" -pattern_type glob -i '/out/review/f*.png' -vf 'scale=640:360,tile=3x3' -frames:v 1 /out/contact.png
find /out -user root -o -type d -name '*.md'
