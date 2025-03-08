package constants

import "time"

// VideoStatus определяет возможные статусы видео
type VideoStatus string

const (
	// VideoStatusUploaded - видео загружено, но еще не в обработке
	VideoStatusUploaded VideoStatus = "uploaded"

	// VideoStatusProcessing - видео загружено и обрабатывается
	VideoStatusProcessing VideoStatus = "processing"

	// VideoStatusReady - видео обработано и готово к просмотру
	VideoStatusReady VideoStatus = "ready"

	// VideoStatusError - ошибка при обработке видео
	VideoStatusError VideoStatus = "error"

	// VideoStatusDeleted - видео удалено
	VideoStatusDeleted VideoStatus = "deleted"
)

const (
	DefaultLimit            = 20
	MaxLimit                = 100
	VideoViewTimeout        = 30 * time.Minute
	VideoCacheDuration      = 1 * time.Hour
	CategoriesCacheDuration = 10 * time.Hour
)
