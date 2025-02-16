package units

import (
	"fmt"
	"strconv"
	"strings"
)

// Code taken from implementation used by Docker.
// For some reason, this module is not exported,
// hence it's copied here.
// See: https://github.com/docker/go-units/blob/master/size_test.go
const (
	// Decimal.
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary.
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

type unitMap map[byte]int64

type ByteSize int64

func (bs *ByteSize) UnmarshalText(data []byte) error {
	size, err := FromHumanSize(string(data))
	if err != nil {
		return fmt.Errorf("parse size %s: %w", string(data), err)
	}

	*bs = ByteSize(size)
	return nil
}

var (
	decimalMap = unitMap{'k': KB, 'm': MB, 'g': GB, 't': TB, 'p': PB}
	binaryMap  = unitMap{'k': KiB, 'm': MiB, 'g': GiB, 't': TiB, 'p': PiB}
)

var (
	decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	binaryAbbrs  = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

func getSizeAndUnit(size float64, base float64, units []string) (float64, string) {
	i := 0
	unitsLimit := len(units) - 1
	for size >= base && i < unitsLimit {
		size /= base
		i++
	}
	return size, units[i]
}

// HumanSizeWithPrecision allows the size to be in any precision,
// instead of 4 digit precision used in units.HumanSize.
func HumanSizeWithPrecision(size float64, precision int) string {
	size, unit := getSizeAndUnit(size, 1000.0, decimapAbbrs)
	return fmt.Sprintf("%.*g%s", precision, size, unit)
}

// HumanSize returns a human-readable approximation of a size
// capped at 4 valid numbers (eg. "2.746 MB", "796 KB").
func HumanSize(size float64) string {
	return HumanSizeWithPrecision(size, 4)
}

// FromHumanSize returns an integer from a human-readable specification of a
// size using SI standard (eg. "44kB", "17MB").
func FromHumanSize(size string) (int64, error) {
	return parseSize(size, decimalMap)
}

// Parses the human-readable size string into the amount it represents.
func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	sep := strings.LastIndexAny(sizeStr, "01234567890. ")
	if sep == -1 {
		// There should be at least a digit.
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}
	var num, sfx string
	if sizeStr[sep] != ' ' {
		num = sizeStr[:sep+1]
		sfx = sizeStr[sep+1:]
	} else {
		// Omit the space separator.
		num = sizeStr[:sep]
		sfx = sizeStr[sep+1:]
	}

	size, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return -1, err
	}
	// Backward compatibility: reject negative sizes.
	if size < 0 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	if len(sfx) == 0 {
		return int64(size), nil
	}

	// Process the suffix.

	if len(sfx) > 3 { // Too long.
		goto badSuffix
	}
	sfx = strings.ToLower(sfx)
	// Trivial case: b suffix.
	if sfx[0] == 'b' {
		if len(sfx) > 1 { // no extra characters allowed after b.
			goto badSuffix
		}
		return int64(size), nil
	}
	// A suffix from the map.
	if mul, ok := uMap[sfx[0]]; ok {
		size *= float64(mul)
	} else {
		goto badSuffix
	}

	// The suffix may have extra "b" or "ib" (e.g. KiB or MB).
	switch {
	case len(sfx) == 2 && sfx[1] != 'b':
		goto badSuffix
	case len(sfx) == 3 && sfx[1:] != "ib":
		goto badSuffix
	}

	return int64(size), nil

badSuffix:
	return -1, fmt.Errorf("invalid suffix: '%s'", sfx)
}
