package memory

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type KV interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Del(key string) error
	MGet(keys []string) ([]string, error)
	MSet(keys, values []string) error
	Keys() ([]string, error)
}

type Store struct {
	kv    KV
	clock func() time.Time
}

func NewStore(kv KV) *Store {
	return &Store{
		kv:    kv,
		clock: time.Now,
	}
}

func (s *Store) PutBasic(m BasicMemory) (string, error) {
	if m.ID == "" {
		m.ID = newID()
	}
	now := s.clock()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	if err := s.kv.Set(basicKey(m.ID), string(data)); err != nil {
		return "", err
	}
	if err := s.updateBasicIndexes(m); err != nil {
		return "", err
	}
	return m.ID, nil
}

func (s *Store) GetBasic(id string) (BasicMemory, error) {
	raw, err := s.kv.Get(basicKey(id))
	if err != nil {
		return BasicMemory{}, err
	}
	var m BasicMemory
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return BasicMemory{}, err
	}
	return m, nil
}

func (s *Store) ListBasic(filter BasicFilter) ([]BasicMemory, error) {
	ids, err := s.basicIDs(filter)
	if err != nil {
		return nil, err
	}
	ids = applyLimit(ids, filter.Limit)
	return s.loadBasics(ids)
}

func (s *Store) DeleteBasic(id string) error {
	m, err := s.GetBasic(id)
	if err != nil {
		return err
	}
	if err := s.kv.Del(basicKey(id)); err != nil {
		return err
	}
	return s.removeBasicIndexes(m)
}

func (s *Store) PutAdvanced(m AdvancedMemory) (string, error) {
	if m.ID == "" {
		m.ID = newID()
	}
	now := s.clock()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	if err := s.kv.Set(advancedKey(m.ID), string(data)); err != nil {
		return "", err
	}
	if err := s.updateAdvancedIndexes(m); err != nil {
		return "", err
	}
	return m.ID, nil
}

func (s *Store) GetAdvanced(id string) (AdvancedMemory, error) {
	raw, err := s.kv.Get(advancedKey(id))
	if err != nil {
		return AdvancedMemory{}, err
	}
	var m AdvancedMemory
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return AdvancedMemory{}, err
	}
	return m, nil
}

func (s *Store) ListAdvanced(filter AdvancedFilter) ([]AdvancedMemory, error) {
	ids, err := s.advancedIDs(filter)
	if err != nil {
		return nil, err
	}
	ids = applyLimit(ids, filter.Limit)
	return s.loadAdvanced(ids)
}

func (s *Store) DeleteAdvanced(id string) error {
	m, err := s.GetAdvanced(id)
	if err != nil {
		return err
	}
	if err := s.kv.Del(advancedKey(id)); err != nil {
		return err
	}
	return s.removeAdvancedIndexes(m)
}

func (s *Store) PutPolicy(m CognitivePolicy) (string, error) {
	if m.ID == "" {
		m.ID = newID()
	}
	now := s.clock()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	if err := s.kv.Set(policyKey(m.ID), string(data)); err != nil {
		return "", err
	}
	if err := s.updatePolicyIndexes(m); err != nil {
		return "", err
	}
	return m.ID, nil
}

func (s *Store) GetPolicy(id string) (CognitivePolicy, error) {
	raw, err := s.kv.Get(policyKey(id))
	if err != nil {
		return CognitivePolicy{}, err
	}
	var m CognitivePolicy
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return CognitivePolicy{}, err
	}
	return m, nil
}

func (s *Store) ListPolicy(filter PolicyFilter) ([]CognitivePolicy, error) {
	ids, err := s.policyIDs(filter)
	if err != nil {
		return nil, err
	}
	ids = applyLimit(ids, filter.Limit)
	return s.loadPolicies(ids)
}

func (s *Store) DeletePolicy(id string) error {
	m, err := s.GetPolicy(id)
	if err != nil {
		return err
	}
	if err := s.kv.Del(policyKey(id)); err != nil {
		return err
	}
	return s.removePolicyIndexes(m)
}

func (s *Store) basicIDs(filter BasicFilter) ([]string, error) {
	switch {
	case filter.Tag != "":
		return s.indexList(basicIndexTag(filter.Tag))
	case filter.Kind != "":
		return s.indexList(basicIndexKind(filter.Kind))
	case filter.Scope != "":
		return s.indexList(basicIndexScope(filter.Scope))
	case filter.Subject != "":
		return s.indexList(basicIndexSubject(filter.Subject))
	default:
		keys, err := s.kv.Keys()
		if err != nil {
			return nil, err
		}
		ids := []string{}
		for _, k := range keys {
			if strings.HasPrefix(k, basicPrefix) {
				ids = append(ids, strings.TrimPrefix(k, basicPrefix))
			}
		}
		sort.Strings(ids)
		return ids, nil
	}
}

func (s *Store) advancedIDs(filter AdvancedFilter) ([]string, error) {
	switch {
	case filter.Tag != "":
		return s.indexList(advancedIndexTag(filter.Tag))
	case filter.Template != "":
		return s.indexList(advancedIndexTemplate(filter.Template))
	default:
		keys, err := s.kv.Keys()
		if err != nil {
			return nil, err
		}
		ids := []string{}
		for _, k := range keys {
			if strings.HasPrefix(k, advancedPrefix) {
				ids = append(ids, strings.TrimPrefix(k, advancedPrefix))
			}
		}
		sort.Strings(ids)
		return ids, nil
	}
}

func (s *Store) policyIDs(filter PolicyFilter) ([]string, error) {
	switch {
	case filter.Tag != "":
		return s.indexList(policyIndexTag(filter.Tag))
	case filter.Role != "":
		return s.indexList(policyIndexRole(filter.Role))
	case filter.Domain != "":
		return s.indexList(policyIndexDomain(filter.Domain))
	case filter.Scenario != "":
		return s.indexList(policyIndexScenario(filter.Scenario))
	default:
		keys, err := s.kv.Keys()
		if err != nil {
			return nil, err
		}
		ids := []string{}
		for _, k := range keys {
			if strings.HasPrefix(k, policyPrefix) {
				ids = append(ids, strings.TrimPrefix(k, policyPrefix))
			}
		}
		sort.Strings(ids)
		return ids, nil
	}
}

func (s *Store) loadBasics(ids []string) ([]BasicMemory, error) {
	if len(ids) == 0 {
		return []BasicMemory{}, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = basicKey(id)
	}
	raws, err := s.kv.MGet(keys)
	if err != nil {
		return nil, err
	}
	out := make([]BasicMemory, 0, len(ids))
	for _, raw := range raws {
		if raw == "" {
			continue
		}
		var m BasicMemory
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) loadAdvanced(ids []string) ([]AdvancedMemory, error) {
	if len(ids) == 0 {
		return []AdvancedMemory{}, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = advancedKey(id)
	}
	raws, err := s.kv.MGet(keys)
	if err != nil {
		return nil, err
	}
	out := make([]AdvancedMemory, 0, len(ids))
	for _, raw := range raws {
		if raw == "" {
			continue
		}
		var m AdvancedMemory
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) loadPolicies(ids []string) ([]CognitivePolicy, error) {
	if len(ids) == 0 {
		return []CognitivePolicy{}, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = policyKey(id)
	}
	raws, err := s.kv.MGet(keys)
	if err != nil {
		return nil, err
	}
	out := make([]CognitivePolicy, 0, len(ids))
	for _, raw := range raws {
		if raw == "" {
			continue
		}
		var m CognitivePolicy
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) updateBasicIndexes(m BasicMemory) error {
	if m.Kind != "" {
		if err := s.indexAdd(basicIndexKind(m.Kind), m.ID); err != nil {
			return err
		}
	}
	if m.Scope != "" {
		if err := s.indexAdd(basicIndexScope(m.Scope), m.ID); err != nil {
			return err
		}
	}
	if m.Subject != "" {
		if err := s.indexAdd(basicIndexSubject(m.Subject), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexAdd(basicIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) removeBasicIndexes(m BasicMemory) error {
	if m.Kind != "" {
		if err := s.indexRemove(basicIndexKind(m.Kind), m.ID); err != nil {
			return err
		}
	}
	if m.Scope != "" {
		if err := s.indexRemove(basicIndexScope(m.Scope), m.ID); err != nil {
			return err
		}
	}
	if m.Subject != "" {
		if err := s.indexRemove(basicIndexSubject(m.Subject), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexRemove(basicIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) updateAdvancedIndexes(m AdvancedMemory) error {
	if m.Template != "" {
		if err := s.indexAdd(advancedIndexTemplate(m.Template), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexAdd(advancedIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) removeAdvancedIndexes(m AdvancedMemory) error {
	if m.Template != "" {
		if err := s.indexRemove(advancedIndexTemplate(m.Template), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexRemove(advancedIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) updatePolicyIndexes(m CognitivePolicy) error {
	if m.Role != "" {
		if err := s.indexAdd(policyIndexRole(m.Role), m.ID); err != nil {
			return err
		}
	}
	if m.Domain != "" {
		if err := s.indexAdd(policyIndexDomain(m.Domain), m.ID); err != nil {
			return err
		}
	}
	if m.Scenario != "" {
		if err := s.indexAdd(policyIndexScenario(m.Scenario), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexAdd(policyIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) removePolicyIndexes(m CognitivePolicy) error {
	if m.Role != "" {
		if err := s.indexRemove(policyIndexRole(m.Role), m.ID); err != nil {
			return err
		}
	}
	if m.Domain != "" {
		if err := s.indexRemove(policyIndexDomain(m.Domain), m.ID); err != nil {
			return err
		}
	}
	if m.Scenario != "" {
		if err := s.indexRemove(policyIndexScenario(m.Scenario), m.ID); err != nil {
			return err
		}
	}
	for _, tag := range m.Tags {
		if err := s.indexRemove(policyIndexTag(tag), m.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) indexAdd(key, id string) error {
	ids, err := s.indexList(key)
	if err != nil {
		return err
	}
	for _, existing := range ids {
		if existing == id {
			return nil
		}
	}
	ids = append(ids, id)
	return s.setIndexList(key, ids)
}

func (s *Store) indexRemove(key, id string) error {
	ids, err := s.indexList(key)
	if err != nil {
		return err
	}
	next := make([]string, 0, len(ids))
	for _, existing := range ids {
		if existing != id {
			next = append(next, existing)
		}
	}
	if len(next) == 0 {
		return s.kv.Del(key)
	}
	return s.setIndexList(key, next)
}

func (s *Store) indexList(key string) ([]string, error) {
	raw, err := s.kv.Get(key)
	if err != nil {
		return []string{}, nil
	}
	if raw == "" {
		return []string{}, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Store) setIndexList(key string, ids []string) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return s.kv.Set(key, string(data))
}

const (
	basicPrefix    = "/_memory/basic/"
	advancedPrefix = "/_memory/advanced/"
	policyPrefix   = "/_memory/policy/"
)

func basicKey(id string) string {
	return basicPrefix + id
}

func advancedKey(id string) string {
	return advancedPrefix + id
}

func policyKey(id string) string {
	return policyPrefix + id
}

func basicIndexKind(kind BasicKind) string {
	return "/_memory/index/basic/kind/" + string(kind)
}

func basicIndexScope(scope string) string {
	return "/_memory/index/basic/scope/" + scope
}

func basicIndexSubject(subject string) string {
	return "/_memory/index/basic/subject/" + subject
}

func basicIndexTag(tag string) string {
	return "/_memory/index/basic/tag/" + tag
}

func advancedIndexTemplate(template string) string {
	return "/_memory/index/advanced/template/" + template
}

func advancedIndexTag(tag string) string {
	return "/_memory/index/advanced/tag/" + tag
}

func policyIndexRole(role string) string {
	return "/_memory/index/policy/role/" + role
}

func policyIndexDomain(domain string) string {
	return "/_memory/index/policy/domain/" + domain
}

func policyIndexScenario(scenario string) string {
	return "/_memory/index/policy/scenario/" + scenario
}

func policyIndexTag(tag string) string {
	return "/_memory/index/policy/tag/" + tag
}

func newID() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return time.Now().UTC().Format("20060102150405.000000000") + "-" + hex.EncodeToString(buf)
}

func applyLimit(ids []string, limit int) []string {
	if limit <= 0 || len(ids) <= limit {
		return ids
	}
	return ids[:limit]
}
