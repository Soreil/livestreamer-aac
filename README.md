Checks a specific channel for stream status (ustream only for now) and starts livestreamer. Output is encoded into an ACC stream, which is served through HTTP. Plays a fallback track on loop, when stream is down.

#Dependacies
* livestreamer
* ffmpeg compiled with libfdk-aac

#Usage
* Use `./start.sh` helper script or just run the binary directly
* Edit `./config.init` to configure
