package core

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

const (
	// CompressionThreshold - files larger than this will be compressed
	CompressionThreshold = 1024 // 1KB
)

// CompressContent compresses content if it's above the threshold
func CompressContent(content string) (string, bool, error) {
	if len(content) < CompressionThreshold {
		return content, false, nil
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write([]byte(content))
	if err != nil {
		return content, false, err
	}

	err = writer.Close()
	if err != nil {
		return content, false, err
	}

	// Encode compressed content as base64
	compressed := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Only use compression if it actually saves space
	if len(compressed) < len(content) {
		return compressed, true, nil
	}

	return content, false, nil
}

// DecompressContent decompresses content if it was compressed
func DecompressContent(content string, compressed bool) (string, error) {
	if !compressed {
		return content, nil
	}

	// Decode from base64
	compressedData, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}
