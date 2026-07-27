## QWEN 3 TTS

qwen_tts -d qwen3-tts-1.7b -l Chinese -s serena -T 0.3 --seed 42   --text "你好" --stdout | ffmpeg -f s16le -ar 24000 -ac 1 -i - -c:a libopus -b:a 32k ~/Downloads/card01.opus