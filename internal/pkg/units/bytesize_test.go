package units

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestHumanSize(t *testing.T) {
	assertEquals(t, "1kB", HumanSize(1000))
	assertEquals(t, "1.024kB", HumanSize(1024))
	assertEquals(t, "1MB", HumanSize(1000000))
	assertEquals(t, "1.049MB", HumanSize(1048576))
	assertEquals(t, "2MB", HumanSize(2*MB))
	assertEquals(t, "3.42GB", HumanSize(float64(3.42*GB)))
	assertEquals(t, "5.372TB", HumanSize(float64(5.372*TB)))
	assertEquals(t, "2.22PB", HumanSize(float64(2.22*PB)))
	assertEquals(t, "1e+04YB", HumanSize(float64(10000000000000*PB)))
}

func TestFromHumanSize(t *testing.T) {
	assertSuccessEquals(t, 0, FromHumanSize, "0")
	assertSuccessEquals(t, 0, FromHumanSize, "0b")
	assertSuccessEquals(t, 0, FromHumanSize, "0B")
	assertSuccessEquals(t, 0, FromHumanSize, "0 B")
	assertSuccessEquals(t, 32, FromHumanSize, "32")
	assertSuccessEquals(t, 32, FromHumanSize, "32b")
	assertSuccessEquals(t, 32, FromHumanSize, "32B")
	assertSuccessEquals(t, 32*KB, FromHumanSize, "32k")
	assertSuccessEquals(t, 32*KB, FromHumanSize, "32K")
	assertSuccessEquals(t, 32*KB, FromHumanSize, "32kb")
	assertSuccessEquals(t, 32*KB, FromHumanSize, "32Kb")
	assertSuccessEquals(t, 32*MB, FromHumanSize, "32Mb")
	assertSuccessEquals(t, 32*GB, FromHumanSize, "32Gb")
	assertSuccessEquals(t, 32*TB, FromHumanSize, "32Tb")
	assertSuccessEquals(t, 32*PB, FromHumanSize, "32Pb")

	assertSuccessEquals(t, 32.5*KB, FromHumanSize, "32.5kB")
	assertSuccessEquals(t, 32.5*KB, FromHumanSize, "32.5 kB")
	assertSuccessEquals(t, 32, FromHumanSize, "32.5 B")
	assertSuccessEquals(t, 300, FromHumanSize, "0.3 K")
	assertSuccessEquals(t, 300, FromHumanSize, ".3kB")

	assertSuccessEquals(t, 0, FromHumanSize, "0.")
	assertSuccessEquals(t, 0, FromHumanSize, "0. ")
	assertSuccessEquals(t, 0, FromHumanSize, "0.b")
	assertSuccessEquals(t, 0, FromHumanSize, "0.B")
	assertSuccessEquals(t, 0, FromHumanSize, "-0")
	assertSuccessEquals(t, 0, FromHumanSize, "-0b")
	assertSuccessEquals(t, 0, FromHumanSize, "-0B")
	assertSuccessEquals(t, 0, FromHumanSize, "-0 b")
	assertSuccessEquals(t, 0, FromHumanSize, "-0 B")
	assertSuccessEquals(t, 32, FromHumanSize, "32.")
	assertSuccessEquals(t, 32, FromHumanSize, "32.b")
	assertSuccessEquals(t, 32, FromHumanSize, "32.B")
	assertSuccessEquals(t, 32, FromHumanSize, "32. b")
	assertSuccessEquals(t, 32, FromHumanSize, "32. B")

	// We do not tolerate extra leading or trailing spaces
	// (except for a space after the number and a missing suffix).
	assertSuccessEquals(t, 0, FromHumanSize, "0 ")

	assertError(t, FromHumanSize, " 0")
	assertError(t, FromHumanSize, " 0b")
	assertError(t, FromHumanSize, " 0B")
	assertError(t, FromHumanSize, " 0 B")
	assertError(t, FromHumanSize, "0b ")
	assertError(t, FromHumanSize, "0B ")
	assertError(t, FromHumanSize, "0 B ")

	assertError(t, FromHumanSize, "")
	assertError(t, FromHumanSize, "hello")
	assertError(t, FromHumanSize, ".")
	assertError(t, FromHumanSize, ". ")
	assertError(t, FromHumanSize, " ")
	assertError(t, FromHumanSize, "  ")
	assertError(t, FromHumanSize, " .")
	assertError(t, FromHumanSize, " . ")
	assertError(t, FromHumanSize, "-32")
	assertError(t, FromHumanSize, "-32b")
	assertError(t, FromHumanSize, "-32B")
	assertError(t, FromHumanSize, "-32 b")
	assertError(t, FromHumanSize, "-32 B")
	assertError(t, FromHumanSize, "32b.")
	assertError(t, FromHumanSize, "32B.")
	assertError(t, FromHumanSize, "32 b.")
	assertError(t, FromHumanSize, "32 B.")
	assertError(t, FromHumanSize, "32 bb")
	assertError(t, FromHumanSize, "32 BB")
	assertError(t, FromHumanSize, "32 b b")
	assertError(t, FromHumanSize, "32 B B")
	assertError(t, FromHumanSize, "32  b")
	assertError(t, FromHumanSize, "32  B")
	assertError(t, FromHumanSize, " 32 ")
	assertError(t, FromHumanSize, "32m b")
	assertError(t, FromHumanSize, "32bm")
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected '%v' but got '%v'", expected, actual)
	}
}

// parseFn - func that maps to the parse function signatures as testing abstraction
type parseFn func(string) (int64, error)

// String used for pretty printing test output.
func (fn parseFn) String() string {
	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return fnName[strings.LastIndex(fnName, ".")+1:]
}

func assertSuccessEquals(t *testing.T, expected int64, fn parseFn, arg string) {
	t.Helper()
	res, err := fn(arg)
	if err != nil || res != expected {
		t.Errorf("%s(\"%s\") -> expected '%d' but got '%d' with error '%v'", fn, arg, expected, res, err)
	}
}

func assertError(t *testing.T, fn parseFn, arg string) {
	t.Helper()
	res, err := fn(arg)
	if err == nil && res != -1 {
		t.Errorf("%s(\"%s\") -> expected error but got '%d'", fn, arg, res)
	}
}
