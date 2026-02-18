package lsm

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	opSet byte = 1
	opDel byte = 2
)

type wal struct {
	path string
	file *os.File
	mu   sync.Mutex
}

func newWAL(dir string) (*wal, error) {
	path := filepath.Join(dir, "wal.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &wal{path: path, file: file}, nil
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
	return err
}

func (w *wal) replay(apply func(entry) error) error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		op, err := reader.ReadByte()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		seqBuf := make([]byte, 8)
		if _, err := io.ReadFull(reader, seqBuf); err != nil {
			return err
		}
		keyLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, keyLenBuf); err != nil {
			return err
		}
		valLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, valLenBuf); err != nil {
			return err
		}
		keyLen := binary.BigEndian.Uint32(keyLenBuf)
		valLen := binary.BigEndian.Uint32(valLenBuf)
		keyBytes := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, keyBytes); err != nil {
			return err
		}
		valBytes := make([]byte, valLen)
		if _, err := io.ReadFull(reader, valBytes); err != nil {
			return err
		}
		e := entry{
			key:     string(keyBytes),
			value:   string(valBytes),
			deleted: op == opDel,
			seq:     binary.BigEndian.Uint64(seqBuf),
		}
		if err := apply(e); err != nil {
			return err
		}
	}
}

func (w *wal) reset() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.file.Close(); err != nil {
		return err
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = file
	return nil
}

func (w *wal) close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
