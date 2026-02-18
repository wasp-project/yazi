package lsm

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	opSet byte = 1
	opDel byte = 2
)

type wal struct {
	dir              string
	currentID        uint64
	file             *os.File
	mu               sync.Mutex
	entriesInSegment int
	maxEntries       int
}

func newWAL(dir string, maxEntries int) (*wal, error) {
	segments, err := listWALSegments(dir)
	if err != nil {
		return nil, err
	}
	var currentID uint64 = 1
	if len(segments) > 0 {
		currentID = segments[len(segments)-1]
	}
	file, err := openWALFile(dir, currentID, os.O_CREATE|os.O_RDWR|os.O_APPEND)
	if err != nil {
		return nil, err
	}
	return &wal{
		dir:              dir,
		currentID:        currentID,
		file:             file,
		maxEntries:       maxEntries,
		entriesInSegment: 0,
	}, nil
}

func (w *wal) appendSet(seq uint64, key, val string) error {
	return w.append(opSet, seq, key, val)
}

func (w *wal) appendDel(seq uint64, key string) error {
	return w.append(opDel, seq, key, "")
}

func (w *wal) append(op byte, seq uint64, key, val string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	keyBytes := []byte(key)
	valBytes := []byte(val)
	buf := make([]byte, 1+8+4+4+len(keyBytes)+len(valBytes))
	buf[0] = op
	binary.BigEndian.PutUint64(buf[1:9], seq)
	binary.BigEndian.PutUint32(buf[9:13], uint32(len(keyBytes)))
	binary.BigEndian.PutUint32(buf[13:17], uint32(len(valBytes)))
	copy(buf[17:17+len(keyBytes)], keyBytes)
	copy(buf[17+len(keyBytes):], valBytes)
	_, err := w.file.Write(buf)
	if err != nil {
		return err
	}
	w.entriesInSegment++
	if w.maxEntries > 0 && w.entriesInSegment >= w.maxEntries {
		return w.rotateLocked()
	}
	return nil
}

func (w *wal) replay(apply func(entry) error) error {
	segments, err := listWALSegments(w.dir)
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		segments = []uint64{w.currentID}
	}
	for _, id := range segments {
		f, err := openWALFile(w.dir, id, os.O_CREATE|os.O_RDONLY)
		if err != nil {
			return err
		}
		reader := bufio.NewReader(f)
		for {
			op, err := reader.ReadByte()
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = f.Close()
				return err
			}
			seqBuf := make([]byte, 8)
			if _, err := io.ReadFull(reader, seqBuf); err != nil {
				_ = f.Close()
				return err
			}
			keyLenBuf := make([]byte, 4)
			if _, err := io.ReadFull(reader, keyLenBuf); err != nil {
				_ = f.Close()
				return err
			}
			valLenBuf := make([]byte, 4)
			if _, err := io.ReadFull(reader, valLenBuf); err != nil {
				_ = f.Close()
				return err
			}
			keyLen := binary.BigEndian.Uint32(keyLenBuf)
			valLen := binary.BigEndian.Uint32(valLenBuf)
			keyBytes := make([]byte, keyLen)
			if _, err := io.ReadFull(reader, keyBytes); err != nil {
				_ = f.Close()
				return err
			}
			valBytes := make([]byte, valLen)
			if _, err := io.ReadFull(reader, valBytes); err != nil {
				_ = f.Close()
				return err
			}
			e := entry{
				key:     string(keyBytes),
				value:   string(valBytes),
				deleted: op == opDel,
				seq:     binary.BigEndian.Uint64(seqBuf),
			}
			if err := apply(e); err != nil {
				_ = f.Close()
				return err
			}
		}
		_ = f.Close()
	}
	return nil
}

func (w *wal) reset() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.resetLocked()
}

func (w *wal) close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func (w *wal) resetLocked() error {
	segments, err := listWALSegments(w.dir)
	if err != nil {
		return err
	}
	if err := w.file.Close(); err != nil {
		return err
	}
	for _, id := range segments {
		_ = os.Remove(walPath(w.dir, id))
	}
	w.currentID++
	file, err := openWALFile(w.dir, w.currentID, os.O_CREATE|os.O_RDWR|os.O_APPEND)
	if err != nil {
		return err
	}
	w.file = file
	w.entriesInSegment = 0
	return nil
}

func (w *wal) rotateLocked() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	w.currentID++
	file, err := openWALFile(w.dir, w.currentID, os.O_CREATE|os.O_RDWR|os.O_APPEND)
	if err != nil {
		return err
	}
	w.file = file
	w.entriesInSegment = 0
	return nil
}

func listWALSegments(dir string) ([]uint64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []uint64{}, nil
		}
		return nil, err
	}
	ids := []uint64{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "wal-") || !strings.HasSuffix(name, ".log") {
			continue
		}
		idStr := strings.TrimSuffix(strings.TrimPrefix(name, "wal-"), ".log")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func walPath(dir string, id uint64) string {
	return filepath.Join(dir, "wal-"+strconv.FormatUint(id, 10)+".log")
}

func openWALFile(dir string, id uint64, flag int) (*os.File, error) {
	return os.OpenFile(walPath(dir, id), flag, 0644)
}
