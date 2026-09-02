package main

import (
	"os"
	"path/filepath"
)

// Storage maps the flat torrent byte-stream onto one or more files on disk.
type Storage struct {
	files       []*storageFile
	totalLength int
}

type storageFile struct {
	f      *os.File
	start  int // global byte offset where this file begins
	length int
}

func NewStorage(outDir string, t *Torrent) (*Storage, error) {
	info := t.Data.Info
	s := &Storage{totalLength: t.downloadLength}

	if !t.multiFile {
		path := filepath.Join(outDir, info.Name)
		f, err := createFile(path, info.Length)
		if err != nil {
			return nil, err
		}
		s.files = append(s.files, &storageFile{f: f, start: 0, length: info.Length})
		return s, nil
	}

	offset := 0
	for _, file := range info.Files {
		parts := append([]string{outDir, info.Name}, file.Path...)
		f, err := createFile(filepath.Join(parts...), file.Length)
		if err != nil {
			return nil, err
		}
		s.files = append(s.files, &storageFile{f: f, start: offset, length: file.Length})
		offset += file.Length
	}
	return s, nil
}

func createFile(path string, length int) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if err := f.Truncate(int64(length)); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// WritePiece writes an assembled, verified piece to the correct offset, spanning
// file boundaries as needed.
func (s *Storage) WritePiece(pieceIndex, pieceLength int, data []byte) error {
	globalStart := pieceIndex * pieceLength
	globalEnd := globalStart + len(data)

	for _, sf := range s.files {
		fileEnd := sf.start + sf.length
		if globalEnd <= sf.start || globalStart >= fileEnd {
			continue // no overlap
		}
		from := max(globalStart, sf.start)
		to := min(globalEnd, fileEnd)
		chunk := data[from-globalStart : to-globalStart]
		if _, err := sf.f.WriteAt(chunk, int64(from-sf.start)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) Close() {
	for _, sf := range s.files {
		sf.f.Close()
	}
}
