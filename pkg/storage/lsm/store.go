package lsm

import (
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wasp-project/yazi/pkg/config"
)

type entry struct {
	key     string
	value   string
	deleted bool
	seq     uint64
}

type Store struct {
	mu                  sync.RWMutex
	mem                 map[string]entry
	wal                 *wal
	sstables            []*sstable
	nextID              uint64
	seq                 uint64
	dir                 string
	memtableMaxEntries  int
	compactionMaxTables int
}

func NewStore(conf config.LSMConfig) (*Store, error) {
	if conf.MemtableMaxEntries <= 0 {
		conf.MemtableMaxEntries = 1024
	}
	if conf.CompactionMaxTables <= 0 {
		conf.CompactionMaxTables = 4
	}
	if conf.Dir == "" {
		conf.Dir = "data/lsm"
	}
	if err := os.MkdirAll(conf.Dir, 0755); err != nil {
		return nil, err
	}
	tables, nextID, maxSeq, err := loadSSTables(conf.Dir)
	if err != nil {
		return nil, err
	}
	w, err := newWAL(conf.Dir, conf.WALMaxSegmentEntries)
	if err != nil {
		return nil, err
	}
	s := &Store{
		mem:                 map[string]entry{},
		wal:                 w,
		sstables:            tables,
		nextID:              nextID,
		seq:                 maxSeq,
		dir:                 conf.Dir,
		memtableMaxEntries:  conf.MemtableMaxEntries,
		compactionMaxTables: conf.CompactionMaxTables,
	}
	if err := s.wal.replay(func(e entry) error {
		if e.seq > s.seq {
			s.seq = e.seq
		}
		if existing, ok := s.mem[e.key]; ok {
			if e.seq <= existing.seq {
				return nil
			}
		}
		s.mem[e.key] = e
		return nil
	}); err != nil {
		return nil, err
	}
	if len(s.mem) >= s.memtableMaxEntries {
		if err := s.flushLocked(); err != nil {
			return nil, err
		}
	}
	if len(s.sstables) >= s.compactionMaxTables {
		if err := s.compactLocked(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Get(key string) (string, error) {
	key = strings.TrimSpace(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.mem[key]; ok {
		if e.deleted {
			return "", errors.New("key not found")
		}
		return e.value, nil
	}
	for i := len(s.sstables) - 1; i >= 0; i-- {
		if e, ok := s.sstables[i].entries[key]; ok {
			if e.deleted {
				return "", errors.New("key not found")
			}
			return e.value, nil
		}
	}
	return "", errors.New("key not found")
}

func (s *Store) Set(key, val string) error {
	key = strings.TrimSpace(key)
	return s.applyWrite(key, val, false)
}

func (s *Store) Expire(key string, ttl time.Duration) error {
	return nil
}

func (s *Store) Del(key string) error {
	key = strings.TrimSpace(key)
	return s.applyWrite(key, "", true)
}

func (s *Store) MSet(keys, values []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range keys {
		key := strings.TrimSpace(keys[i])
		val := strings.TrimSpace(values[i])
		if err := s.applyWriteLocked(key, val, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) MGet(keys []string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]string, len(keys))
	for i := range keys {
		key := strings.TrimSpace(keys[i])
		if e, ok := s.mem[key]; ok {
			if e.deleted {
				continue
			}
			values[i] = e.value
			continue
		}
		for j := len(s.sstables) - 1; j >= 0; j-- {
			if e, ok := s.sstables[j].entries[key]; ok {
				if e.deleted {
					break
				}
				values[i] = e.value
				break
			}
		}
	}
	return values, nil
}

func (s *Store) Keys() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := map[string]entry{}
	for k, v := range s.mem {
		all[k] = v
	}
	for i := len(s.sstables) - 1; i >= 0; i-- {
		for k, v := range s.sstables[i].entries {
			if _, ok := all[k]; ok {
				continue
			}
			all[k] = v
		}
	}
	keys := make([]string, 0, len(all))
	for k, v := range all {
		if v.deleted {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func (s *Store) Encode() []byte {
	return nil
}

func (s *Store) Decode(data []byte) error {
	return nil
}

func (s *Store) Close() error {
	return s.wal.close()
}

func (s *Store) applyWrite(key, val string, deleted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyWriteLocked(key, val, deleted)
}

func (s *Store) applyWriteLocked(key, val string, deleted bool) error {
	s.seq++
	seq := s.seq
	var err error
	if deleted {
		err = s.wal.appendDel(seq, key)
	} else {
		err = s.wal.appendSet(seq, key, val)
	}
	if err != nil {
		s.seq--
		return err
	}
	s.mem[key] = entry{
		key:     key,
		value:   val,
		deleted: deleted,
		seq:     seq,
	}
	if len(s.mem) >= s.memtableMaxEntries {
		if err := s.flushLocked(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) flushLocked() error {
	if len(s.mem) == 0 {
		return nil
	}
	entries := make([]entry, 0, len(s.mem))
	for _, e := range s.mem {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	table, err := writeSSTable(s.dir, s.nextID, entries)
	if err != nil {
		return err
	}
	s.sstables = append(s.sstables, table)
	s.nextID++
	s.mem = map[string]entry{}
	if err := s.wal.reset(); err != nil {
		return err
	}
	if len(s.sstables) >= s.compactionMaxTables {
		return s.compactLocked()
	}
	return nil
}

func (s *Store) compactLocked() error {
	if len(s.sstables) <= 1 {
		return nil
	}
	merged := map[string]entry{}
	for i := 0; i < len(s.sstables); i++ {
		for k, v := range s.sstables[i].entries {
			existing, ok := merged[k]
			if !ok || v.seq > existing.seq {
				merged[k] = v
			}
		}
	}
	entries := make([]entry, 0, len(merged))
	for _, e := range merged {
		if e.deleted {
			continue
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	table, err := writeSSTable(s.dir, s.nextID, entries)
	if err != nil {
		return err
	}
	for _, t := range s.sstables {
		_ = os.Remove(t.path)
	}
	s.sstables = []*sstable{table}
	s.nextID++
	return nil
}
