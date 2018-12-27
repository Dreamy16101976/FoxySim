package libs

//go get github.com/gonum/matrix
//go get github.com/gonum/floats

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math"
	"math/cmplx"
	"strconv"
	"strings"
)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Str2Val(nominal string) (float64, bool) {
	var power float64 = 1.0
	var value float64
	var err error
	powerCh := nominal[len(nominal)-1]
	switch powerCh {
	case 'K':
		{
			power = 1E3
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'M':
		{
			power = 1E6
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'G':
		{
			power = 1E9
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'T':
		{
			power = 1E12
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'm':
		{
			power = 1E-3
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'u':
		{
			power = 1E-6
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'n':
		{
			power = 1E-9
			nominal = nominal[0 : len(nominal)-1]
		}
	case 'p':
		{
			power = 1E-12
			nominal = nominal[0 : len(nominal)-1]
		}

	}

	value, err = strconv.ParseFloat(nominal, 64)
	if err != nil {
		return 0.0, false
	}
	value = value * power
	return value, true
}

//возвращает значение в радианах
func Str2Angle(angle string) (float64, bool) {
	var mul = math.Pi / 180.0
	unitCh := angle[len(angle)-1]
	switch unitCh {
	case 'd':
		{
			mul = math.Pi / 180.0
			angle = angle[0 : len(angle)-1]
		}
	case 'r':
		{
			mul = 1.0
			angle = angle[0 : len(angle)-1]
		}
	}
	value, success := Str2Val(angle)
	if success == false {
		return 0.0, false
	}
	value = value * mul
	return value, true
}

//возвращает значение угловой частоты в рад/с
func Str2Freq(freq string) (float64, bool) {
	var mul = 2.0 * math.Pi
	unitCh := freq[len(freq)-1]
	switch unitCh {
	case 'f':
		{
			mul = 2.0 * math.Pi
			freq = freq[0 : len(freq)-1]
		}
	case 'w':
		{
			mul = 1.0
			freq = freq[0 : len(freq)-1]
		}
	}
	value, success := Str2Val(freq)
	if success == false {
		return 0.0, false
	}
	value = value * mul
	return value, true
}

func Float2Str(value float64, decfmt string, digits int) string {
	tmp := ""
	if decfmt == "SCI" {
		tmp = strconv.FormatFloat(value, 'E', digits, 64)
	} else {
		tmp = strconv.FormatFloat(value, 'f', digits, 64)
		tmp = strings.TrimRight(tmp, "0")
		tmp = strings.TrimRight(tmp, ".")
	}
	return tmp
}

func Complex2Str(value complex128, angle string, decfmt string, digits int) string {
	absTmp := ""
	if decfmt == "SCI" {
		absTmp = strconv.FormatFloat(cmplx.Abs(value), 'E', digits, 64)
	} else {
		absTmp = strconv.FormatFloat(cmplx.Abs(value), 'f', digits, 64)
		absTmp = strings.TrimRight(absTmp, "0")
		absTmp = strings.TrimRight(absTmp, ".")
	}
	res := absTmp
	if angle == "RAD" { //радианы
		argTmp := strconv.FormatFloat(cmplx.Phase(value), 'f', 3, 64)
		argTmp = strings.TrimRight(argTmp, "0")
		argTmp = strings.TrimRight(argTmp, ".")
		res = res + "<br/>&ang; " + argTmp
	} else { //градусы
		argTmp := strconv.FormatFloat(cmplx.Phase(value)*180.0/math.Pi, 'f', 3, 64)
		argTmp = strings.TrimRight(argTmp, "0")
		argTmp = strings.TrimRight(argTmp, ".")
		res = res + "<br/>&ang; " + argTmp + "&deg;"
	}
	return res
}

func Crt1DFloat(m int) []float64 {
	a := make([]float64, m)
	return a
}

func Crt1DComplex(m int) []complex128 {
	a := make([]complex128, m)
	return a
}

func Crt2DFloat(m, n int) [][]float64 {
	a := make([][]float64, m)
	for i := 0; i < m; i++ {
		a[i] = make([]float64, n)
	}
	return a
}

func Crt2DComplex(m, n int) [][]complex128 {
	a := make([][]complex128, m)
	for i := 0; i < m; i++ {
		a[i] = make([]complex128, n)
	}
	return a
}

func T2DFloat(a [][]float64) [][]float64 {
	m := len(a)
	n := len(a[0])
	c := Crt2DFloat(n, m)
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			c[i][j] = a[j][i]
		}
	}
	return c
}

func T2DComplex(a [][]complex128) [][]complex128 {
	m := len(a)
	n := len(a[0])
	c := Crt2DComplex(n, m)
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			c[i][j] = a[j][i]
		}
	}
	return c
}

func Mul2DComplex1DComplex(a [][]complex128, b []complex128) ([]complex128, bool) {
	m := len(a)
	n := len(a[0])
	c := Crt1DComplex(m)
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			c[j] = c[j] + a[j][i]*b[i]
		}
	}
	return c, true
}

func Mul2DFloat1DComplex(a [][]float64, b []complex128) ([]complex128, bool) {
	n := len(a)
	m := len(a[0])
	c := Crt1DComplex(n)
	for i := 0; i < n; i++ {
		c[i] = 0 + 0i
		for k := 0; k < m; k++ {
			c[i] = c[i] + cmplx.Rect(a[i][k], 0)*b[k]
		}
	}
	return c, true
}

func Mul2DFloat2DComplex(a [][]float64, b [][]complex128) ([][]complex128, bool) {
	n := len(a)
	m := len(a[0])
	p := len(b[0])
	c := Crt2DComplex(n, p)
	for i := 0; i < n; i++ {
		for j := 0; j < p; j++ {
			c[i][j] = 0 + 0i
			for k := 0; k < m; k++ {
				c[i][j] = c[i][j] + cmplx.Rect(a[i][k], 0)*b[k][j]
			}
		}
	}
	return c, true
}

func Mul2DComplex2DComplex(a [][]complex128, b [][]complex128) ([][]complex128, bool) {
	n := len(a)
	m := len(a[0])
	p := len(b[0])
	c := Crt2DComplex(n, p)
	for i := 0; i < n; i++ {
		for j := 0; j < p; j++ {
			c[i][j] = 0 + 0i
			for k := 0; k < m; k++ {
				c[i][j] = c[i][j] + a[i][k]*b[k][j]
			}
		}
	}
	return c, true
}

func Mul2DComplex2DFloat(a [][]complex128, b [][]float64) ([][]complex128, bool) {
	n := len(a)
	m := len(a[0])
	p := len(b[0])
	c := Crt2DComplex(n, p)
	for i := 0; i < n; i++ {
		for j := 0; j < p; j++ {
			c[i][j] = 0 + 0i
			for k := 0; k < m; k++ {
				c[i][j] = c[i][j] + a[i][k]*cmplx.Rect(b[k][j], 0)
			}
		}
	}
	return c, true
}

func GaussComplex(a0 [][]complex128, b0 []complex128) ([]complex128, error) {
	// make augmented matrix
	m := len(b0)
	a := make([][]complex128, m)
	for i, ai := range a0 {
		row := make([]complex128, m+1)
		copy(row, ai)
		row[m] = b0[i]
		a[i] = row
	}
	// WP algorithm from Gaussian elimination page
	// produces row-eschelon form
	for k := range a {
		// Find pivot for column k:
		iMax := k
		max := cmplx.Abs(a[k][k])
		for i := k + 1; i < m; i++ {
			if abs := cmplx.Abs(a[i][k]); abs > max {
				iMax = i
				max = abs
			}
		}
		if a[iMax][k] == 0 {
			return nil, errors.New("singular")
		}
		// swap rows(k, i_max)
		a[k], a[iMax] = a[iMax], a[k]
		// Do for all rows below pivot:
		for i := k + 1; i < m; i++ {
			// Do for all remaining elements in current row:
			for j := k + 1; j <= m; j++ {
				a[i][j] -= a[k][j] * (a[i][k] / a[k][k])
			}
			// Fill lower triangular matrix with zeros:
			a[i][k] = 0
		}
	}
	// end of WP algorithm.
	// now back substitute to get result.
	x := make([]complex128, m)
	for i := m - 1; i >= 0; i-- {
		x[i] = a[i][m]
		for j := i + 1; j < m; j++ {
			x[i] -= a[i][j] * x[j]
		}
		x[i] /= a[i][i]
	}
	return x, nil
}
