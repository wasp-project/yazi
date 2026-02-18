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
)

type sstable struct {
	id      uint64
	path    string
	entries map[string]entry
}

func loadSSTables(dir string) ([]*sstable, uint64, uint64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*sstable{}, 1, 0, nil
		}
		return nil, 0, 0, err
	}
	type sstableFile struct {
		id   uint64
		path string
	}
	var files []sstableFile
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "sstable-") || !strings.HasSuffix(name, ".dat") {
			continue
		}
		idStr := strings.TrimSuffix(strings.TrimPrefix(name, "sstable-"), ".dat")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		files = append(files, sstableFile{id: id, path: filepath.Join(dir, name)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].id < files[j].id })
	var maxID uint64
	var maxSeq uint64
	var tables []*sstable
	for _, f := range files {
		table, seq, err := loadSSTable(f.path, f.id)
		if err != nil {
			return nil, 0, 0, err
		}
		tables = append(tables, table)
		if f.id > maxID {
			maxID = f.id
		}
		if seq > maxSeq {
			maxSeq = seq
		}
	}
	nextID := maxID + 1
	if nextID == 0 {
		nextID = 1
	}
	return tables, nextID, maxSeq, nil
}

func loadSSTable(path string, id uint64) (*sstable, uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	entries := map[string]entry{}
	var maxSeq uint64
	for {
		seqBuf := make([]byte, 8)
		if _, err := io.ReadFull(reader, seqBuf); err != nil {
			if err == io.EOF {
				break
			}
			if err == io.ErrUnexpectedEOF {
				break
			}
			return nil, 0, err
		}
		delByte, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, err
		}
		keyLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, keyLenBuf); err != nil {
			return nil, 0, err
		}
		valLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, valLenBuf); err != nil {
			return nil, 0, err
		}
		keyLen := binary.BigEndian.Uint32(keyLenBuf)
		valLen := binary.BigEndian.Uint32(valLenBuf)
		keyBytes := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, keyBytes); err != nil {
			return nil, 0, err
		}
		valBytes := make([]byte, valLen)
		if _, err := io.ReadFull(reader, valBytes); err != nil {
			return nil, 0, err
		}
		seq := binary.BigEndian.Uint64(seqBuf)
		if seq > maxSeq {
			maxSeq = seq
		}
		entries[string(keyBytes)] = entry{
			key:     string(keyBytes),
			value:   string(valBytes),
			deleted: delByte == 1,
			seq:     seq,
		}
	}
	return &sstable{id: id, path: path, entries: entries}, maxSeq, nil
}

func writeSSTable(dir string, id uint64, entries []entry) (*sstable, error) {
	name := filepath.Join(dir, "sstable-"+strconv.FormatUint(id, 10)+".dat")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, e := range entries {
		seqBuf := make([]byte, 8)
		binary.BigEndian.PutUint64(seqBuf, e.seq)
		if _, err := writer.Write(seqBuf); err != nil {
			return nil, err
		}
		if e.deleted {
			if err := writer.WriteByte(1); err != nil {
				return nil, err
			}
		} else {
			if err := writer.WriteByte(0); err != nil {
				return nil, err
			}
		}
		keyBytes := []byte(e.key)
		valBytes := []byte(e.value)
		keyLenBuf := make([]byte, 4)
		valLenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(keyLenBuf, uint32(len(keyBytes)))
		binary.BigEndian.PutUint32(valLenBuf, uint32(len(valBytes)))
		if _, err := writer.Write(keyLenBuf); err != nil {
			return nil, err
		}
		if _, err := writer.Write(valLenBuf); err != nil {
			return nil, err
		}
		if _, err := writer.Write(keyBytes); err != nil {
			return nil, err
		}
		if _, err := writer.Write(valBytes); err != nil {
			return nil, err
		}
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	table := &sstable{
		id:      id,
		path:    name,
		entries: map[string]entry{},
	}
	for _, e := range entries {
		table.entries[e.key] = e
	}
	return table, nil
}
