package testing

import (
	"bytes"
	"fmt"
	"regexp"
	tt "testing"
)

// CheckErrMatches checks whether the given error matches the regular
// expression given as "match".
func CheckErrMatches(t *tt.T, err error, match string) bool {
	t.Helper()

	switch {
	case match == "" && err == nil:
		return true

	case err == nil:
		t.Errorf("expected error matching %s, but got nil", match)

	case match == "":
		t.Errorf("expected nil error, but got %#v", err)

	case !regexp.MustCompile(match).MatchString(err.Error()):
		t.Errorf("expected error matching %s, but got %#v",
			match, err)
	default:
		return true
	}

	return false
}

// CheckEq tests two values for equality.
func CheckEq(t *tt.T, given, expect interface{}) bool {
	t.Helper()

	if given != expect {
		t.Errorf("expected %#v to equal %#v", given, expect)
		return false
	}
	return true
}

// CheckBufEq tests two byte slices for equality.
func CheckBufEq(t *tt.T, given, expect []byte) bool {
	t.Helper()

	if !bytes.Equal(given, expect) {
		t.Errorf("expected:\n%#q\nbut got:\n%#q\n",
			renderBytes(given),
			renderBytes(expect),
		)

		// TODO: t.Errorf("diff: %s", renderDiff(given, expect))
		return false
	}

	return true
}

// func renderDiff(a, b []byte) string {
// 	var (
// 		diffs [][]byte
// 		offs []int
// 		inDiff bool
// 		thisDiff []byte
// 	)

// 	for i := 0; i < len(a) && i < len(b); i++ {
// 		switch {
// 		case inDiff && a[i] == b[i]:
// 			inDiff = false
// 			diffs = append(diffs, thisDiff)
// 		case !inDiff && a[i] != b[i]:
// 			inDiff = true
// 			thisDiff = []byte{a[i]}
// 			if !inDiff {
// 				inDiff = true
// 			if thisDiff
// 	}
// }

func renderBytes(some []byte) string {
	// TODO: Hex dump with ASCII on the right
	ls := len(some)
	if ls < 72 {
		return string(some)
	}

	return fmt.Sprintf("%s... (%d more)", some[:59], ls-59)
}
