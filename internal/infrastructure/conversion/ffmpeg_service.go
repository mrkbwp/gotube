package conversion

import (
	"fmt"
	"github.com/mrkbwp/gotube/internal/domain/entity"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type FFmpegService struct {
	tempDir string
}

func NewFFmpegService(tempDir string) *FFmpegService {
	return &FFmpegService{
		tempDir: tempDir,
	}
}

func (s *FFmpegService) ConvertVideo(inputPath string, outputPath string, quality *entity.VideoQuality) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-b:v", fmt.Sprintf("%dk", quality.Bitrate),
		"-vf", fmt.Sprintf("scale=%d:%d", quality.Width, quality.Height),
		"-c:a", "aac",
		"-y",
		outputPath,
	)

	return cmd.Run()
}

func (s *FFmpegService) GetVideoInfo(inputPath string) (duration int, err error) {
	log.Printf("Getting video info for: %s", inputPath)

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("ffprobe command failed: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("ffprobe stderr: %s", string(exitErr.Stderr))
		}
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	log.Printf("Raw duration output: '%s'", durationStr)

	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		log.Printf("Failed to parse duration string: %v", err)
		return 0, fmt.Errorf("failed to parse duration '%s': %w", durationStr, err)
	}

	duration = int(durationFloat)
	log.Printf("Parsed duration: %d seconds", duration)

	return duration, nil
}

func (s *FFmpegService) GenerateThumbnail(inputPath string, outputPath string) error {
	// Сначала получаем длительность видео
	duration, err := s.GetVideoInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	// Берем кадр из середины видео
	middleTime := duration / 2

	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%d", middleTime),
		"-i", inputPath,
		"-vframes", "1",
		"-q:v", "2",
		"-f", "image2",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return nil
}
