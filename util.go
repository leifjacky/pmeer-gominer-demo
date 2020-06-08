package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/Qitmeer/crypto/sha3"
	"golang.org/x/crypto/blake2b"
)

func UInt32LEToBytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return b
}

func UInt64BEToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func MustStringToHexBytes(st string) []byte {
	b, _ := hex.DecodeString(st)
	return b
}

func Hash2BigTarget(hash []byte) *big.Int {
	return new(big.Int).SetBytes(hash[:])
}

func MustParseInt64(str string, base int) int64 {
	i, _ := strconv.ParseInt(str, base, 64)
	return i
}

func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("util: Can't parse duration `" + s + "`: " + err.Error())
	}
	return value
}

func GetReadableHashRateString(hashrate float64) string {
	if hashrate <= 0 {
		return "0 " + "H"
	}

	units := []string{"H", "K", "M", "G", "T", "P", "E", "Z", "Y"}

	i := int64(math.Min(float64(len(units)-1), math.Max(0.0, math.Floor(math.Log(hashrate)/math.Log(1000.0)))))
	hr_float := hashrate / math.Pow(1000.0, float64(i))

	return fmt.Sprintf("%.3f %s", hr_float, units[i])
}

func FillZeroHashLen(hash string, l int) string {
	for len(hash) < l {
		hash = "0" + hash
	}
	return hash
}

var (
	DiffOneTarget      = new(big.Int).Sub(new(big.Int).Lsh(BigOne, 224), BigOne)
	DiffOneTargetFloat = new(big.Float).SetInt(DiffOneTarget)
)

func Diff2BigTarget(diff float64) *big.Int {
	if diff <= 0 {
		return new(big.Int).SetInt64(0)
	}
	i := new(big.Int)
	new(big.Float).Quo(DiffOneTargetFloat, new(big.Float).SetFloat64(diff)).Int(i)
	return i
}

func BigTarget2Diff(target *big.Int) float64 {
	diff, _ := new(big.Float).Quo(new(big.Float).SetInt(DiffOneTarget), new(big.Float).SetInt(target)).Float64()
	return diff
}

func Blake2bD(data []byte) []byte {
	return Blake2b(Blake2b(data))
}

func Blake2b(data []byte) []byte {
	output := blake2b.Sum256(data)
	return output[:]
}

// Reverse reverses a byte array.
func ReverseBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

func ReverseStringByte(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-2; from < to; from, to = from+2, to-2 {
		runes[from], runes[to] = runes[to], runes[from]
		runes[from+1], runes[to+1] = runes[to+1], runes[from+1]
	}

	return string(runes)
}

var (
	HashFunc = Blake2bD
)

func SetHashFunc(f func([]byte) []byte) {
	HashFunc = f
}

func QitmeerKeecak256(data []byte) []byte {
	keccak := sha3.NewQitmeerKeccak256()
	keccak.Write(data)
	r := keccak.Sum(nil)
	return r
}
