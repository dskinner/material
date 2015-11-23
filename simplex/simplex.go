package simplex

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"time"
)

// Interesting: https://www.stat.washington.edu/research/reports/1991/tr201.pdf

type Vec []float32

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

func (m Mat) Column(j int) (v Vec) {
	for _, row := range m {
		v = append(v, row[j])
	}
	return
}

func (m Mat) Transpose(a *Mat) {
	*a = make(Mat, len(m[0]))
	for j := range *a {
		(*a)[j] = m.Column(j)
	}
}

func (m Mat) String() string {
	var buf bytes.Buffer
	for _, row := range m {
		buf.WriteString(fmt.Sprintf("% 4.2f\n", row))
	}
	return buf.String()
}

type Var struct {
	int
	Val float32
}

type Coef struct {
	C float32
	V Var
}

type Relation int

const (
	GreaterEq Relation = iota - 1
	Equal
	LessEq
)

type Constraint struct {
	Vs []Coef
	R  Relation
	C  float32
}

func Constrain(vs ...Coef) Constraint {
	return Constraint{Vs: vs}
}

func (cn Constraint) LessEq(x float32) Constraint {
	cn.R, cn.C = LessEq, x
	return cn
}

func (cn Constraint) GreaterEq(x float32) Constraint {
	cn.R, cn.C = GreaterEq, x
	return cn
}

func (cn Constraint) Equal(x float32) Constraint {
	cn.R, cn.C = Equal, x
	return cn
}

type Program struct {
	c Vec // coefficients
	a Mat // constraints
	b Vec // equalities
	s Mat // surplus/slack

	tbl Mat
}

func (prg *Program) Var(coef float32) Var {
	prg.c = append(prg.c, coef)
	return Var{len(prg.c) - 1, 0}
}

func (prg *Program) AddConstraints(cns ...Constraint) {
	for _, cn := range cns {
		vcn := make(Vec, len(prg.c))
		for _, v := range cn.Vs {
			vcn[v.V.int] = v.C
		}
		prg.a = append(prg.a, vcn)
		prg.b = append(prg.b, cn.C)

		switch cn.R {
		case GreaterEq, LessEq, Equal: // TODO avoid unnecessary Equal when gen'ing tableau
			for i := range prg.s {
				prg.s[i] = append(prg.s[i], 0)
			}
			s := make([]float32, len(prg.b))
			s[len(s)-1] = float32(cn.R)
			prg.s = append(prg.s, s)
		// case Equal:
		default:
			panic("unknown relation")
		}
	}
}

func (prg *Program) Tableau() Mat {
	// TODO this is all pretty cheesy
	tbl := make(Mat, len(prg.a)+1)
	c := make([]float32, len(prg.c))
	for i, x := range prg.c {
		c[i] = -x
	}
	tbl[0] = c
	copy(tbl[1:], prg.a)

	tbl.Transpose(&tbl)
	// ident := make([]float32, len(tbl[0]))
	// ident[0] = 1
	// tbl.Insert(0, ident)

	s := make(Mat, len(prg.s))
	copy(s, prg.s)
	s.Transpose(&s)
	for _, v := range s {
		v.Insert(0, 0)
		tbl = append(tbl, v)
	}

	b := make([]float32, len(prg.b)+1)
	copy(b[1:], prg.b)
	tbl = append(tbl, b)
	tbl.Transpose(&tbl)
	return tbl
}

func (prg *Program) Minimize() error {
	t := time.Now()
	prg.tbl = prg.Tableau()
	// fmt.Printf("t:\n%s", prg.tbl)
	// fmt.Println("---")

	var err error
	var i, j int
	for i, j, err = PivotIndex(prg.tbl); err == nil; i, j, err = PivotIndex(prg.tbl) {
		// fmt.Printf("[i=%v][j=%v]\n", i, j)
		Pivot(prg.tbl, i, j)
		// fmt.Printf("%s", prg.tbl)
		// fmt.Println("---")
	}

	// tbl.Transpose(&tbl)
	// fmt.Printf("%s", tbl)
	// fmt.Println("---")
	if err != nil {
		log.Println(err)
	}
	log.Printf("simplex: finished in %s\n", time.Now().Sub(t))
	return nil
}

func (prg *Program) For(vars ...*Var) {
	j := len(prg.tbl[0]) - 1
	for _, v := range vars {
		for i, x := range prg.tbl.Column((*v).int) {
			if x == 1 {
				(*v).Val = prg.tbl[i][j]
				break
			}
		}
	}
}

func PivotIndex(tbl Mat) (int, int, error) {
	pj := -1
	// for j, x := range tbl[0][:len(tbl[0])-1] {
	tmp := tbl[0][:len(tbl[0])-1]
	for j := len(tmp) - 1; j >= 0; j-- {
		x := tmp[j]
		// TODO Note that by changing the entering variable choice rule so that it selects a column where the
		// entry in the objective row is negative, the algorithm is changed so that it finds the maximum
		// of the objective function rather than the minimum.
		// TODO Picking the largest available element as the pivot is usually a good choice
		// http://mathworld.wolfram.com/Gauss-JordanElimination.html
		if x < 0 { // see Note above; `<` is maximize, `>` is minimize
			pj = j
			break
		}
	}
	if pj == -1 {
		return 0, 0, fmt.Errorf("simplex: optimal solution reached")
	}

	// TODO If there is more than one row for which the minimum is achieved
	// then a dropping variable choice rule can be used to make the determination.
	var pi = -1
	var min float32 = math.MaxFloat32
	for i, r := range tbl[1:] {
		i++ // real index
		if r[pj] > 0 {
			if n := r[len(r)-1] / r[pj]; n < min {
				min = n
				pi = i
			}
		}
	}
	if pi == -1 {
		return 0, 0, fmt.Errorf("simplex: function is unbounded")
	}

	return pi, pj, nil
}

func Pivot(tbl Mat, pi, pj int) {
	div := tbl[pi][pj]
	for j := range tbl[pi] {
		tbl[pi][j] /= div
	}
	for i := range tbl {
		if i != pi {
			var dys []float32
			for _, y := range tbl[pi] {
				dys = append(dys, y*tbl[i][pj])
			}
			for j, dy := range dys {
				tbl[i][j] -= dy
			}
		}
	}
}
