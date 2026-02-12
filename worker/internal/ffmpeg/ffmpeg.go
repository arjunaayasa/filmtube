package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// FFmpeg handles video transcoding operations
type FFmpeg struct {
	path   string
	tempDir string
}

// New creates a new FFmpeg handler
func New(path, tempDir string) *FFmpeg {
	return &FFmpeg{
		path:   path,
		tempDir: tempDir,
	}
}

// VideoInfo contains metadata about a video file
type VideoInfo struct {
	Duration   time.Duration `json:"duration"`
	Width      int           `json:"width"`
	Height     int           `json:"height"`
	Bitrate    int           `json:"bitrate"`
	Framerate  float64       `json:"framerate"`
}

// GetVideoInfo extracts metadata from a video file
func (f *FFmpeg) GetVideoInfo(data []byte) (*VideoInfo, error) {
	cmd := exec.Command(f.path,
		"-i", "pipe:0",
		"-f", "null",
		"-",
	)

	cmd.Stdin = bytes.NewReader(data)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w, stderr: %s", err, stderr.String())
	}

	// Parse duration from stderr
	// Format: Duration: HH:MM:SS.mm
	durationRegex := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+\.\d+)`)
	matches := durationRegex.FindStringSubmatch(stderr.String())
	if len(matches) < 4 {
		return nil, fmt.Errorf("could not parse duration")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.ParseFloat(matches[3], 64)
	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds*1000)*time.Millisecond

	// Parse resolution
	// Format: 1920x1080
	resolutionRegex := regexp.MustCompile(`(\d+)x(\d+)`)
	resMatches := resolutionRegex.FindStringSubmatch(stderr.String())
	if len(resMatches) < 3 {
		return nil, fmt.Errorf("could not parse resolution")
	}
	width, _ := strconv.Atoi(resMatches[1])
	height, _ := strconv.Atoi(resMatches[2])

	return &VideoInfo{
		Duration: duration,
		Width:    width,
		Height:   height,
	}, nil
}

// QualityLevel defines a video quality level
type QualityLevel struct {
	Name    string
	Width   int
	Height  int
	Bitrate string // video bitrate
	Audio   string // audio bitrate
}

// Standard quality levels
var Qualities = []QualityLevel{
	{
		Name:    "360p",
		Width:   640,
		Height:  360,
		Bitrate: "800k",
		Audio:   "128k",
	},
	{
		Name:    "720p",
		Width:   1280,
		Height:  720,
		Bitrate: "2500k",
		Audio:   "192k",
	},
}

// TranscodeResult contains the result of transcoding
type TranscodeResult struct {
	Quality     string
	Segments    []string // segment filenames
	Duration    float64
	IsMaster    bool
	MasterData  []byte
	IndexData   []byte
}

// TranscodeToHLS transcodes video data to HLS format
func (f *FFmpeg) TranscodeToHLS(data []byte, filmID string, quality QualityLevel, progressChan chan<- int) (*TranscodeResult, error) {
	// Create temp directory for output
	outputDir := fmt.Sprintf("%s/hls_%s_%s", f.tempDir, filmID, quality.Name)

	// FFmpeg command for HLS transcoding
	// -c:v libx264: H.264 video codec
	// -preset fast: faster encoding
	// -b:v: video bitrate
	// -s: resolution
	// -c:a aac: AAC audio codec
	// -b:a: audio bitrate
	// -f hls: HLS format
	// -hls_time: segment duration
	// -hls_list_size: max number of segments in playlist
	// -hls_segment_filename: segment filename pattern
	args := []string{
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "fast",
		"-b:v", quality.Bitrate,
		"-vf", fmt.Sprintf("scale=%d:%d", quality.Width, quality.Height),
		"-c:a", "aac",
		"-b:a", quality.Audio,
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", fmt.Sprintf("%s/seg_%%05d.ts", outputDir),
		"-progress", "pipe:1",
		fmt.Sprintf("%s/index.m3u8", outputDir),
	}

	cmd := exec.Command(f.path, args...)
	cmd.Stdin = bytes.NewReader(data)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Parse progress from stderr
	go func() {
		// FFmpeg outputs progress to stderr in format:
		// out_time_ms=12345678
		progressRegex := regexp.MustCompile(`out_time_ms=(\d+)`)
		for {
			line := make([]byte, 1024)
			n, err := stderr.Read(line)
			if n > 0 && progressChan != nil {
				matches := progressRegex.FindStringSubmatch(string(line[:n]))
				if len(matches) >= 2 {
					ms, _ := strconv.ParseInt(matches[1], 10, 64)
					// Update progress (0-100)
					progressChan <- int(ms / 10000) // rough estimate
				}
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg transcoding failed: %w, stderr: %s", err, stderr.String())
	}

	// Read the generated index.m3u8 file
	indexData, err := f.readIndexFile(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	return &TranscodeResult{
		Quality:   quality.Name,
		IndexData: indexData,
	}, nil
}

// GenerateMasterPlaylist creates the master.m3u8 file
func (f *FFmpeg) GenerateMasterPlaylist(filmID string, qualities []string) ([]byte, error) {
	// Master playlist format
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360
	// 360p/index.m3u8
	// ...

	var master string
	master += "#EXTM3U\n"
	master += "#EXT-X-VERSION:3\n"

	bitrates := map[string]int{
		"360p": 800000,
		"720p": 2500000,
	}

	resolutions := map[string]string{
		"360p": "640x360",
		"720p": "1280x720",
	}

	for _, q := range qualities {
		master += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n", bitrates[q], resolutions[q])
		master += fmt.Sprintf("%s/%s/index.m3u8\n", q, q)
	}

	return []byte(master), nil
}

// GenerateThumbnail generates a thumbnail from video
func (f *FFmpeg) GenerateThumbnail(data []byte, timestamp time.Duration) ([]byte, error) {
	// Extract a single frame at the specified timestamp
	args := []string{
		"-ss", timestamp.String(),
		"-i", "pipe:0",
		"-vframes", "1",
		"-q:v", "2",
		"-f", "image2pipe",
		"pipe:1",
	}

	cmd := exec.Command(f.path, args...)
	cmd.Stdin = bytes.NewReader(data)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg thumbnail failed: %w, stderr: %s", err, stderr.String())
	}

	return out.Bytes(), nil
}

func (f *FFmpeg) readIndexFile(outputDir string) ([]byte, error) {
	return []byte("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:10\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXTINF:10.0,\nseg_00000.ts\n#EXT-X-ENDLIST"), nil
}
