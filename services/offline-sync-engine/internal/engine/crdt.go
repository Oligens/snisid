package engine

import (
	"encoding/json"
)

type ConflictResult int

const (
	BEFORE ConflictResult = iota
	AFTER
	CONCURRENT
	EQUAL
)

type VectorClock map[string]int

func (vc VectorClock) Increment(nodeID string) {
	vc[nodeID]++
}

func (vc VectorClock) Compare(other VectorClock) ConflictResult {
	if len(vc) == 0 && len(other) == 0 {
		return EQUAL
	}
	if len(vc) == 0 {
		return BEFORE
	}
	if len(other) == 0 {
		return AFTER
	}

	allKeys := make(map[string]struct{})
	for k := range vc {
		allKeys[k] = struct{}{}
	}
	for k := range other {
		allKeys[k] = struct{}{}
	}

	var vcGreater, otherGreater bool

	for k := range allKeys {
		v1 := vc[k]
		v2 := other[k]
		if v1 > v2 {
			vcGreater = true
		} else if v2 > v1 {
			otherGreater = true
		}
	}

	if vcGreater && !otherGreater {
		return AFTER
	}
	if otherGreater && !vcGreater {
		return BEFORE
	}
	if vcGreater && otherGreater {
		return CONCURRENT
	}
	return EQUAL
}

func (vc VectorClock) Merge(other VectorClock) {
	for k, v := range other {
		if existing, ok := vc[k]; !ok || v > existing {
			vc[k] = v
		}
	}
}

func (vc VectorClock) Serialize() string {
	if vc == nil {
		return "{}"
	}
	b, err := json.Marshal(vc)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func DeserializeVectorClock(data string) VectorClock {
	vc := make(VectorClock)
	if data == "" || data == "{}" {
		return vc
	}
	if err := json.Unmarshal([]byte(data), &vc); err != nil {
		return make(VectorClock)
	}
	return vc
}
