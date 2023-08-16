package util

import (
	"contact-discovery/pir"
	"log"
	"os"

	"github.com/dgryski/go-metro"
	"github.com/linvon/cuckoo-filter"
)

const BitsPerItem = 32
const TagsPerBucket = 3
const BucketLen = TagsPerBucket * BitsPerItem / 8
const TagMask = (1 << BitsPerItem) - 1

func CFToFile(cf *cuckoo.Filter, filename string) {
	b_cf, err := cf.Encode()
	if err != nil {
		log.Fatalln("Error encoding Cuckoo Filter!\n", err)
	}
	err = os.WriteFile(filename, b_cf, 0777)
	if err != nil {
		log.Fatalln("Error writing CF to file \n", err)
	}
}

func CFfromFile(filename string) *cuckoo.Filter {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalln("Error reading file!")
	}

	cf, err2 := cuckoo.Decode(data)
	if err2 != nil {
		log.Fatalln("Error decoding Cuckoo Filter!")
	}

	return cf
}

func CreateCF(elements *[]pir.Row) (cf *cuckoo.Filter) {
	// Create CF
	cf = cuckoo.NewFilter(TagsPerBucket, BitsPerItem,
		uint(len(*elements)), cuckoo.TableTypeSingle)

	for i := 0; i < len(*elements); i++ {
		res := cf.Add((*elements)[i])
		if !res {
			log.Fatalln("Insertion into CF failed!")
		}
	}
	return cf
}

// bucketLen = 12
func CFSegment(cf *cuckoo.Filter, numSeg uint32, segSize uint32) *pir.StaticDB {
	buckets := cf.BucketsSlice(uint64(numSeg*segSize), uint64(segSize))
	ShrinkCapacity(&buckets, len(buckets))
	return &(pir.StaticDB{NumRows: len(buckets) / BucketLen, RowLen: BucketLen, FlatDb: buckets})
}

// Generate tag from hash.
// Adapted from cuckoo-filter lib
func tagHash(hv uint32) uint32 {
	return hv%((1<<BitsPerItem)-1) + 1
}

// Generate index from hash.
// Adapted from cuckoo-filter lib
func indexHash(hv, numBuckets uint32) uint {
	// table.NumBuckets is always a power of two, so modulo can be replaced with bitwise-and:
	return uint(hv) & uint(numBuckets-1)
}

// Generate alt index from index and tag.
// Adapted from cuckoo-filter lib
func altIndex(index uint, tag uint32, numBuckets uint32) uint {
	// 0x5bd1e995 is the hash constant from MurmurHash2
	return indexHash(uint32(index)^(tag*0x5bd1e995), numBuckets)
}

// Get CF hash, indices, tag
// Adapted from cuckoo-filter lib
func GenerateIndicesTagHash(item []byte, numBuckets uint32) (hash uint64, index1 uint, index2 uint, tag uint32) {
	hash = metro.Hash64(item, 1337)
	tag = tagHash(uint32(hash))
	index1 = indexHash(uint32(hash>>32), numBuckets)
	index2 = altIndex(index1, tag, numBuckets)
	return
}

func SlotContains(bucket []byte, tag uint32) bool {
	tag_pir := uint32(bucket[0]) | uint32(bucket[1])<<8 | uint32(bucket[2])<<16 | uint32(bucket[3])<<24
	if tag_pir&TagMask == tag {
		return true
	}
	return false
}
