package simplex

import (
	"bytes"
	"fmt"
)

type Vec []float32

func (a Vec) Dot(b Vec) (c float32) {
	if len(a) != len(b) {
		panic("a Vec length does not match b Vec length")
	}
	for i, x := range a {
		c += x * b[i]
	}
	return
}

func (a Vec) Sub(c *Vec, b Vec) {
	if len(a) != len(b) {
		panic("a Vec length does not match b Vec length")
	}
	for i, x := range a {
		(*c)[i] = x - b[i]
	}
}

func (a Vec) MulScalar(b *Vec, x float32) {
	for i, v := range a {
		(*b)[i] = v * x
	}
}

func (a Vec) Div(b Vec) (c Vec) {
	if len(a) != len(b) {
		panic("a Vec length does not match b Vec length")
	}
	for i, x := range a {
		c = append(c, x/b[i])
	}
	return
}

func (a Vec) Max() (i int, x float32) {
	for j, y := range a {
		if j == 0 || y > x {
			i, x = j, y
		}
	}
	return
}

func (v *Vec) Insert(i int, x float32) {
	*v = append(*v, 0)
	copy((*v)[i+1:], (*v)[i:])
	(*v)[i] = x
}

type Mat []Vec

func (m *Mat) Insert(i int, xs []float32) {
	*m = append(*m, nil)
	copy((*m)[i+1:], (*m)[i:])
	(*m)[i] = xs
}

func (m Mat) Column(v *Vec, j int) {
	n := len(*v)
	for i, row := range m {
		if i == n {
			break
		}
		(*v)[i] = row[j]
	}
}

func (m Mat) ColumnAlloc(j int) (v Vec) {
	for _, row := range m {
		v = append(v, row[j])
	}
	return
}

func (m Mat) Transpose(a *Mat) {
	*a = make(Mat, len(m[0]))
	for j := range *a {
		(*a)[j] = m.ColumnAlloc(j)
	}
}

func (a Mat) MulVec(b Vec) (c Vec) {
	if len(a[0]) != len(b) {
		panic("row length of a Math does not equal b Vec length")
	}
	for _, r := range a {
		c = append(c, r.Dot(b))
	}
	return
}

func (a Mat) IsColIdent(j int) bool {
	var ok bool
	for _, row := range a {
		y := row[j]
		if y == 1 && !ok {
			ok = true
		} else if y != 0 {
			return false
		}
	}
	return ok
}

func (m Mat) String() string {
	var buf bytes.Buffer
	for _, row := range m {
		buf.WriteString(fmt.Sprintf("% 4.2f\n", row))
	}
	n := buf.Len() - 1
	if n < 0 {
		n = 0
	}
	buf.Truncate(n)
	return buf.String()
}
