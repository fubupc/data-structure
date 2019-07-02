package hamt

import (
	"math/rand"
	"testing"

	gart "github.com/plar/go-adaptive-radix-tree"
)

func genTestKVs(n int, max int64) ([]Key, []Value) {
	keys := []Key{}
	vals := []Value{}
	keyMap := make(map[Key]bool)

	for len(keys) < n {
		k := Key(rand.Int63n(max))
		if !keyMap[k] {
			keyMap[k] = true
			keys = append(keys, k)
			vals = append(vals, Value(len(keys)))
		}
	}

	return keys, vals
}

func makeHAMT(keys []Key, vals []Value) *Map {
	m := NewMap()
	for i := 0; i < len(keys); i++ {
		m.Add(&keys[i], &vals[i])
	}
	return m
}

func makeStdMap(keys []Key, vals []Value) map[int64]int64 {
	m := make(map[int64]int64)
	for i := 0; i < len(keys); i++ {
		m[int64(keys[i])] = int64(vals[i])
	}
	return m
}

func makePtrStdMap(keys []Key, vals []Value) map[*Key]*Value {
	m := make(map[*Key]*Value)
	for i := 0; i < len(keys); i++ {
		m[&keys[i]] = &vals[i]
	}
	return m
}

var testKeys, testVals = genTestKVs(1e5, 1e7)
var testHAMT = makeHAMT(testKeys, testVals)
var testStdMap = makeStdMap(testKeys, testVals)
var testPtrStdMap = makePtrStdMap(testKeys, testVals)

func gartKey(k int64) []byte {
	b := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		b[i] = byte(k >> (8 * i) & 0xFF)
	}
	return b
}
func toGARTKey(keys []Key) []gart.Key {
	gk := make([]gart.Key, 0, 64)
	for _, key := range keys {
		gk = append(gk, gartKey(int64(key)))
	}
	return gk
}
func makeGART(keys []gart.Key, vals []Value) gart.Tree {
	t := gart.New()
	for i := 0; i < len(keys); i++ {
		t.Insert(keys[i], vals[i])
	}
	return t
}

var gartKeys = toGARTKey(testKeys)
var testGART = makeGART(gartKeys, testVals)

//func toARTKey(keys []Key) []art.Key {
//	ak := make([]art.Key, 0, len(keys))
//	for _, key := range keys {
//		ak = append(ak, art.Int64Key(int64(key)))
//	}
//	return ak
//}
//func makeART(keys []art.Key, vals []Value) art.ART {
//	t := art.MakeART()
//	for i := 0; i < len(keys); i++ {
//		t = t.Insert(keys[i], unsafe.Pointer(&vals[i]))
//	}
//	return t
//}
//var artKeys = toARTKey(testKeys)
//var testART = makeART(artKeys, testVals)

// ===== Symbol bit width = 5 =====
// test#1:
//BenchmarkHAMT_Add-8         	     100	  15255057 ns/op
//BenchmarkStdMap_Add-8       	     200	   9137427 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  10903658 ns/op
//BenchmarkHAMT_Find-8        	     500	   2880576 ns/op
//BenchmarkStdMap_Find-8      	     500	   3179046 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3208222 ns/op
//
// test#2:
//BenchmarkHAMT_Add-8         	     100	  15367111 ns/op
//BenchmarkStdMap_Add-8       	     200	   9196946 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  10683513 ns/op
//BenchmarkHAMT_Find-8        	     500	   2857944 ns/op
//BenchmarkStdMap_Find-8      	     500	   3192465 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3143876 ns/op

// ===== Symbol bit width = 6 =====
// test#1:
//BenchmarkHAMT_Add-8         	     100	  16304648 ns/op
//BenchmarkStdMap_Add-8       	     200	   9007895 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  10504392 ns/op
//BenchmarkHAMT_Find-8        	     500	   2629162 ns/op
//BenchmarkStdMap_Find-8      	     500	   3313128 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3302081 ns/op
//
// test#2:
//BenchmarkHAMT_Add-8         	     100	  16408295 ns/op
//BenchmarkStdMap_Add-8       	     200	   9050502 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  10587110 ns/op
//BenchmarkHAMT_Find-8        	     500	   2601488 ns/op
//BenchmarkStdMap_Find-8      	     500	   3191200 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3201509 ns/op

func BenchmarkHAMT_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeHAMT(testKeys, testVals)
	}
}

//func BenchmarkART_Add(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		_ = makeART(artKeys, testVals)
//	}
//}

func BenchmarkGART_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeGART(gartKeys, testVals)
	}
}

func BenchmarkStdMap_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeStdMap(testKeys, testVals)
	}
}

func BenchmarkPtrStdMap_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makePtrStdMap(testKeys, testVals)
	}
}

func BenchmarkHAMT_Find(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for k := 0; k < len(testKeys); k++ {
			_ = testHAMT.Find(&testKeys[k])
		}
	}
}

//func BenchmarkART_Find(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		for k := 0; k < len(artKeys); k++ {
//			_ = testART.Search(artKeys[k])
//		}
//	}
//}

func BenchmarkGART_Find(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for k := 0; k < len(gartKeys); k++ {
			_, _ = testGART.Search(gartKeys[k])
		}
	}
}

func BenchmarkStdMap_Find(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for k := 0; k < len(testKeys); k++ {
			_ = testStdMap[int64(testKeys[k])]
		}
	}
}

func BenchmarkPtrStdMap_Find(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for k := 0; k < len(testKeys); k++ {
			_ = testPtrStdMap[&testKeys[k]]
		}
	}
}
