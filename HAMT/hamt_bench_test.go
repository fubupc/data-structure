package hamt

import (
	"math/rand"
	"testing"
)

func genTestKVs(n int) ([]Key, []Value) {
	keys := []Key{}
	vals := []Value{}
	keyMap := make(map[Key]bool)

	for len(keys) < n {
		k := Key(rand.Uint64())
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

//func makeART(keys []Key, vals []Value) art.ART {
//	t := art.MakeART()
//	for i := 0; i < len(keys); i++ {
//		t = t.Insert(art.Int64Key(int64(keys[i])), unsafe.Pointer(&vals[i]))
//	}
//	return t
//}

var testKeys, testVals = genTestKVs(100000)
var testHAMT = makeHAMT(testKeys, testVals)
var testStdMap = makeStdMap(testKeys, testVals)
var testPtrStdMap = makePtrStdMap(testKeys, testVals)

//var testART = makeART(testKeys, testVals)

// Test#1:
//BenchmarkHAMT_Add-8         	     100	  18062655 ns/op
//BenchmarkART_Add-8          	       5	 262592804 ns/op
//BenchmarkStdMap_Add-8       	     100	  10237082 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  11453917 ns/op
//BenchmarkHAMT_Find-8        	     500	   2909724 ns/op
//BenchmarkART_Find-8         	     100	  13153139 ns/op
//BenchmarkStdMap_Find-8      	     500	   3192401 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3273871 ns/op

// Test#2:
//BenchmarkHAMT_Add-8         	     100	  18130498 ns/op
//BenchmarkART_Add-8          	       5	 265791594 ns/op
//BenchmarkStdMap_Add-8       	     100	  10044649 ns/op
//BenchmarkPtrStdMap_Add-8    	     100	  11448546 ns/op
//BenchmarkHAMT_Find-8        	     500	   2813681 ns/op
//BenchmarkART_Find-8         	     100	  13021862 ns/op
//BenchmarkStdMap_Find-8      	     500	   3125170 ns/op
//BenchmarkPtrStdMap_Find-8   	     500	   3108714 ns/op

func BenchmarkHAMT_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeHAMT(testKeys, testVals)
	}
}

//func BenchmarkART_Add(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		_ = makeART(testKeys, testVals)
//	}
//}

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
//		for k := 0; k < len(testKeys); k++ {
//			_ = testART.Search(art.Int64Key(int64(testKeys[k])))
//		}
//	}
//}

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
