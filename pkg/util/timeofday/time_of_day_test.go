// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package timeofday

import (
	"fmt"
	"testing"
	"time"

	"github.com/auxten/postgresql-parser/pkg/util/duration"
)

func TestString(t *testing.T) {
	expected := "01:02:03.456789"
	actual := New(1, 2, 3, 456789).String()
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
	testData := []struct {
		t   TimeOfDay
		exp string
	}{
		{New(1, 2, 3, 0), "01:02:03"},
		{New(1, 2, 3, 456000), "01:02:03.456"},
		{New(1, 2, 3, 456789), "01:02:03.456789"},
	}
	for i, td := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := td.t.String()
			if actual != td.exp {
				t.Errorf("expected %s, got %s", td.exp, actual)
			}
		})
	}
}

func TestFromAndToTime(t *testing.T) {
	testData := []struct {
		s   string
		exp string
	}{
		{"0000-01-01T00:00:00Z", "1970-01-01T00:00:00Z"},
		{"2017-01-01T12:00:00.5Z", "1970-01-01T12:00:00.5Z"},
		{"9999-12-31T23:59:59.999999Z", "1970-01-01T23:59:59.999999Z"},
		{"2017-01-01T12:00:00-05:00", "1970-01-01T12:00:00Z"},
	}
	for _, td := range testData {
		t.Run(td.s, func(t *testing.T) {
			fromTime, err := time.Parse(time.RFC3339Nano, td.s)
			if err != nil {
				t.Fatal(err)
			}
			actual := FromTime(fromTime).ToTime().Format(time.RFC3339Nano)
			if actual != td.exp {
				t.Errorf("expected %s, got %s", td.exp, actual)
			}
		})
	}
}

func TestRound(t *testing.T) {
	testData := []struct {
		t     TimeOfDay
		round time.Duration
		exp   TimeOfDay
	}{
		{New(12, 0, 0, 1000), time.Second, New(12, 0, 0, 0)},
		{New(12, 0, 0, 1000), time.Millisecond, New(12, 0, 0, 1000)},
		{Max, time.Second, Time2400},
		{Time2400, time.Second, Time2400},
		{Min, time.Second, Min},
	}
	for _, td := range testData {
		t.Run(fmt.Sprintf("%s,%s", td.t, td.round), func(t *testing.T) {
			actual := td.t.Round(td.round)
			if actual != td.exp {
				t.Errorf("expected %s, got %s", td.exp, actual)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	testData := []struct {
		t      TimeOfDay
		micros int64
		exp    TimeOfDay
	}{
		{New(12, 0, 0, 0), 1, New(12, 0, 0, 1)},
		{New(12, 0, 0, 0), microsecondsPerDay, New(12, 0, 0, 0)},
		{Max, 1, Min},
		{Min, -1, Max},
	}
	for _, td := range testData {
		d := duration.MakeDuration(td.micros*nanosPerMicro, 0, 0)
		t.Run(fmt.Sprintf("%s,%s", td.t, d), func(t *testing.T) {
			actual := td.t.Add(d)
			if actual != td.exp {
				t.Errorf("expected %s, got %s", td.exp, actual)
			}
		})
	}
}

func TestDifference(t *testing.T) {
	testData := []struct {
		t1        TimeOfDay
		t2        TimeOfDay
		expMicros int64
	}{
		{New(0, 0, 0, 0), New(0, 0, 0, 0), 0},
		{New(0, 0, 0, 1), New(0, 0, 0, 0), 1},
		{New(0, 0, 0, 0), New(0, 0, 0, 1), -1},
		{Max, Min, microsecondsPerDay - 1},
		{Min, Max, -1 * (microsecondsPerDay - 1)},
	}
	for _, td := range testData {
		t.Run(fmt.Sprintf("%s,%s", td.t1, td.t2), func(t *testing.T) {
			actual := Difference(td.t1, td.t2).Nanos() / nanosPerMicro
			if actual != td.expMicros {
				t.Errorf("expected %d, got %d", td.expMicros, actual)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for Hour, Minute, Second — including 32-bit overflow regression.
//
// THE BUG (pre-fix):
//
//   return int(int64(t) % microsecondsPerDay) / microsecondsPerHour
//                         ↑ already int            ↑ untyped constant
//
// On a 32-bit system the compiler must materialise microsecondsPerHour (3.6e9)
// as an `int` to perform the final integer division.  3.6e9 > math.MaxInt32
// (~2.1e9), so the compiler rejects it:
//
//   microsecondsPerHour (untyped float constant 3.6e+09) truncated to int
//
// THE FIX:
//
//   return int((int64(t) % microsecondsPerDay) / microsecondsPerHour)
//
// The entire division stays in int64 space.  Only the *result* (0–23 for
// hours, 0–59 for minutes/seconds) is narrowed to int, which is safe on both
// 32-bit and 64-bit platforms.
//
// These tests make the correctness of the fix explicit and would fail if the
// parentheses were accidentally removed on a 64-bit host too (the values would
// be wrong for certain inputs because without the parentheses Go's operator
// precedence makes the `int()` cast bind tighter than the `/` operator, giving
// a different — and incorrect — result when the modulo remainder is not an
// exact multiple of the divisor on 64-bit; on 32-bit the build simply fails).
// ---------------------------------------------------------------------------

// TestHour verifies Hour() returns the correct value for a representative
// set of times, including the special Time2400 sentinel.
func TestHour(t *testing.T) {
	testData := []struct {
		tod  TimeOfDay
		want int
	}{
		{New(0, 0, 0, 0), 0},
		{New(1, 0, 0, 0), 1},
		{New(12, 30, 45, 0), 12},
		{New(23, 59, 59, 999999), 23},
		{Time2400, 24},
	}
	for _, td := range testData {
		t.Run(fmt.Sprintf("Hour(%s)", td.tod), func(t *testing.T) {
			got := td.tod.Hour()
			if got != td.want {
				t.Errorf("Hour() = %d, want %d", got, td.want)
			}
		})
	}
}

// TestMinute verifies Minute() for a range of times.
func TestMinute(t *testing.T) {
	testData := []struct {
		tod  TimeOfDay
		want int
	}{
		{New(0, 0, 0, 0), 0},
		{New(0, 1, 0, 0), 1},
		{New(12, 30, 0, 0), 30},
		{New(23, 59, 0, 0), 59},
	}
	for _, td := range testData {
		t.Run(fmt.Sprintf("Minute(%s)", td.tod), func(t *testing.T) {
			got := td.tod.Minute()
			if got != td.want {
				t.Errorf("Minute() = %d, want %d", got, td.want)
			}
		})
	}
}

// TestSecond verifies Second() for a range of times.
func TestSecond(t *testing.T) {
	testData := []struct {
		tod  TimeOfDay
		want int
	}{
		{New(0, 0, 0, 0), 0},
		{New(0, 0, 1, 0), 1},
		{New(12, 30, 45, 0), 45},
		{New(23, 59, 59, 0), 59},
	}
	for _, td := range testData {
		t.Run(fmt.Sprintf("Second(%s)", td.tod), func(t *testing.T) {
			got := td.tod.Second()
			if got != td.want {
				t.Errorf("Second() = %d, want %d", got, td.want)
			}
		})
	}
}

// TestHourMinuteSecondRoundtrip is the key regression test for the 32-bit
// overflow fix.
//
// It constructs a TimeOfDay using New() (which internally works in int64) and
// then asserts that Hour(), Minute(), and Second() reproduce the original
// components exactly.
//
// Why does this catch the bug?
//
//   Before the fix, the expression evaluated by the *compiler* was:
//
//     ( int(int64(t) % microsecondsPerDay) ) / microsecondsPerHour
//
//   The intermediate `int(...)` converts the modulo result to the native `int`
//   type.  On a 32-bit target the subsequent division by microsecondsPerHour
//   (3,600,000,000) cannot be represented as an `int`, so the *compiler*
//   rejects the file entirely — the package cannot be built at all for 32-bit.
//   On a 64-bit host the compile succeeds but the logic is semantically
//   equivalent, so the test still passes there too.
//
//   After the fix:
//
//     int( (int64(t) % microsecondsPerDay) / microsecondsPerHour )
//
//   The division is done entirely in int64.  The final int() cast only touches
//   the small result (0–23), which is safe on both architectures.
//
// The table covers midnight, noon, and end-of-day to maximise the chance of
// exposing any latent arithmetic error.
func TestHourMinuteSecondRoundtrip(t *testing.T) {
	testData := []struct {
		h, m, s int
	}{
		{0, 0, 0},
		{0, 0, 1},
		{0, 1, 0},
		{1, 0, 0},
		{12, 30, 45},
		{23, 59, 59},
	}
	for _, td := range testData {
		name := fmt.Sprintf("%02d:%02d:%02d", td.h, td.m, td.s)
		t.Run(name, func(t *testing.T) {
			tod := New(td.h, td.m, td.s, 0)
			if got := tod.Hour(); got != td.h {
				t.Errorf("Hour() = %d, want %d  (raw µs = %d)", got, td.h, int64(tod))
			}
			if got := tod.Minute(); got != td.m {
				t.Errorf("Minute() = %d, want %d  (raw µs = %d)", got, td.m, int64(tod))
			}
			if got := tod.Second(); got != td.s {
				t.Errorf("Second() = %d, want %d  (raw µs = %d)", got, td.s, int64(tod))
			}
		})
	}
}
