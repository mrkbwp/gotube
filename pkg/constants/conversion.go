package constants

import "time"

// Качество видео
const (

	// TODO: вынести в конфиг
	MaxConcurrentConversions = 5
	ConversionCheckInterval  = 30 * time.Second

	VideoQuality240p     = "240p"
	VideoQuality480p     = "480p"
	VideoQuality720p     = "720p"
	VideoQuality1080p    = "1080p"
	VideoQuality4k       = "4k"
	VideoQualityOriginal = "original"
)
