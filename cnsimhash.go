package cnsimhash

import (
	"hash/fnv"

	"github.com/dgryski/go-bits"
	"github.com/wangbin/jiebago/analyse"
)

const (
	WORDS_ALL = -1
)

var (
	extracter analyse.TagExtracter
)

type hashWeigth struct {
	Hash   uint64
	Weight float64
}

// To Load jieba, idf, stop words dictionaries
func LoadDictionary(jiebapath, idfpath, stopwords, synonympath string) error {
	if err := extracter.LoadIdf(idfpath); err != nil {
		return err
	}

	if err := extracter.LoadDictionary(jiebapath); err != nil {
		return err
	}

	if err := extracter.LoadStopWords(stopwords); err != nil {
		return err
	}
	/*if err := extracter.LoadSynonyms(synonympath); err != nil {
		return err
	}
	extracter.CalculateSynonymsMaxIDF()*/
	return nil
}

// calculate unicode simhash with top n keywords  calculate with all words if topN < 0
func UnicodeSimhash(s string, topN int) (uint64, analyse.Segments, []string) {
	if s == "" {
		return 0, nil, nil
	}

	hashes, weightWords, words := extractHash(s, topN)
	if len(hashes) == 0 {
		return 0, weightWords, words
	}

	weights := calWeights(hashes)
	return fingerprint(weights), weightWords, words
}

func hasher(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func extractHash(s string, topN int) ([]hashWeigth, analyse.Segments, []string) {
	weightWords, words := extracter.CNExtractTags(s, topN)
	wordsLen := len(weightWords)
	if wordsLen == 0 {
		return []hashWeigth{}, weightWords, words
	}

	result := make([]hashWeigth, wordsLen)
	for i, w := range weightWords {
		hash := hasher(w.Text())
		result[i] = hashWeigth{hash, w.Weight()}
	}
	return result, weightWords, words
}

func calWeights(hashes []hashWeigth) [64]float64 {
	var weights [64]float64
	for _, v := range hashes {
		for i := uint8(0); i < 64; i++ {
			weight := v.Weight
			if (1 << i & v.Hash) == 0 {
				weight *= -1
			}
			weights[i] += weight
		}
	}
	return weights
}

func fingerprint(weights [64]float64) uint64 {
	var f uint64
	for i := uint8(0); i < 64; i++ {
		if weights[i] >= 0.0 {
			f |= (1 << i)
		}
	}
	return f
}

// Compare calculates the Hamming distance between two 64-bit integers
//
// Currently, this is calculated using the Kernighan method [1]. Other methods
// exist which may be more efficient and are worth exploring at some point
func Compare(a uint64, b uint64) uint8 {
	v := a ^ b
	var c uint8
	for c = 0; v != 0; c++ {
		v &= v - 1
	}
	return c
}

func Distance(v1 uint64, v2 uint64) int {
	return int(bits.Popcnt(v1 ^ v2))
}
