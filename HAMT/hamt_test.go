package hamt

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitMap(t *testing.T) {
	m := bitmap(0)

	for i := uint64(0); i < cardinality; i++ {
		m = m.set(i)
		assert.Equal(t, int(i), m.countBelow(i))
		for j := uint64(0); j < cardinality; j++ {
			if j <= i {
				assert.True(t, m.isSet(j))
			} else {
				assert.False(t, m.isSet(j))
			}
		}
	}
}

func TestNewMap(t *testing.T) {
	m := NewMap()

	keys := []Key{}
	vals := []Value{}
	keyMap := make(map[Key]bool)

	for len(keys) < 100000 {
		k := Key(rand.Uint64())
		if !keyMap[k] {
			keyMap[k] = true
			keys = append(keys, k)
			vals = append(vals, Value(len(keys)))
		}
	}

	for i := 0; i < len(keys); i++ {
		m.Add(&keys[i], &vals[i])
	}

	assert.Equal(t, len(keyMap), m.Count())

	for i := 0; i < len(keys); i++ {
		v := m.Find(&keys[i])
		assert.Equal(t, &vals[i], v, "key=%d val=%d", keys[i], vals[i])
	}
}
