package ociclient

import (
	"os"
)

func determineChunkSize(filePath string) (int64, error) {
	const (
		smallFileSize   = 50 * 1024 * 1024       // 50 MB
		mediumFileSize  = 500 * 1024 * 1024      // 500 MB
		largeFileSize   = 5 * 1024 * 1024 * 1024 // 5 GB
		smallChunkSize  = 5 * 1024 * 1024        // 5 MB
		mediumChunkSize = 10 * 1024 * 1024       // 10 MB
		largeChunkSize  = 50 * 1024 * 1024       // 50 MB
		hugeChunkSize   = 100 * 1024 * 1024      // 100 MB
	)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	fileSize := fileInfo.Size()

	switch {
	case fileSize <= smallFileSize:
		return smallChunkSize, nil
	case fileSize <= mediumFileSize:
		return mediumChunkSize, nil
	case fileSize <= largeFileSize:
		return largeChunkSize, nil
	default:
		return hugeChunkSize, nil
	}
}
