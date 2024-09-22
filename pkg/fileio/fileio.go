package fileio

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func HumanBytes[T Number](b T) string {
	const units = "KMGTE"
	for i := len(units) - 1; i >= 0; i-- {
		if u := int64(1 << ((i + 1) * 10)); int64(b) >= u {
			return fmt.Sprintf("%0.2f%c", MathDiv(b, u, 2), units[i])
		}
	}
	return fmt.Sprintf("%dB", int64(b))
}

func HumanDuration[T Number](d T) string {
	td := time.Duration(d)
	switch {
	case td >= time.Hour:
		return strconv.FormatFloat(MathDiv(d, time.Hour, 2), 'f', 2, 32) + "h"
	case td > time.Minute:
		return strconv.FormatFloat(MathDiv(d, time.Minute, 2), 'f', 2, 32) + "m"
	case td >= time.Second:
		return strconv.FormatFloat(MathDiv(d, time.Second, 3), 'f', 2, 32) + "s"
	case td >= time.Millisecond:
		return strconv.FormatFloat(MathDiv(d, time.Millisecond, 3), 'f', 2, 32) + "ms"
	case td >= time.Microsecond:
		return strconv.FormatFloat(MathDiv(d, time.Microsecond, 3), 'f', 2, 32) + "µs"
	case td >= time.Nanosecond:
		return strconv.Itoa(int(d)) + "ns"
	default:
		return ""
	}
}

func MathDiv[A, B Number](a A, b B, precision ...int) float64 {
	if b <= 0 {
		return 0
	}

	if len(precision) == 0 || precision[0] < 0 {
		return float64(a) / float64(b)
	}

	if precision[0] == 0 {
		return math.Floor(float64(a) / float64(b))
	}

	prec := math.Pow(10, float64(precision[0]))
	return math.Floor(float64(a)*prec/float64(b)) / prec
}
