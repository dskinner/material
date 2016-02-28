package simplex

import "log"

type Eta struct {
	d Vec
	j int
}

type Etas []Eta

// Solve against identity for entering variable index e.
func (es Etas) Solve(ident Mat, a Vec) Vec {
	// For example:
	// column as row of below
	// [  1  0  0  0 ]
	// [  0  1  0  0 ]
	// [ d1 d2 d3 d4 ]
	// [  0  0  0  1 ]
	//
	// [ 1 0 d1 0] [X1] [b1]
	// [ 0 1 d2 0] [X2]=[b2]
	// [ 0 0 d3 0] [X3] [b3]
	// [ 0 0 d4 1] [X4] [b4]
	// To solve:
	// Set X3= b3/d3.
	// for i not equal to 3, Xi = bi - di * X3.

	// b = [ 0 0 8 ]
	// j = 2
	// d = [ 1 3 5 ]
	// y = [ 0 0 8/5 ] // want

	// [ 1.0, 0.0, 1.0 ] [x0]   [0]
	// [ 0.0, 1.0, 3.0 ] [x1] = [0]
	// [ 0.0, 0.0, 5.0 ] [x2]   [8]
	//
	// x0 = 0/1 + 0/0 + 8/0 = 0
	// x1 = 0/0 + 0/1 + 8/0 = 0
	// x2 = (b2 - x0*1 - x1*3)/5
	// !!!!!!!! yes ^^^^

	// b = [ 0 6 8 ]
	// j = 1
	// d = [ 4/5 7/5 1/5 ]
	// want
	// u = [ 0 22/7 8] = [ 0 3.1428 8]

	//    x0   x1   x2
	// [ 1.0, 0.8, 0.0 ] [x0]   [0]
	// [ 0.0, 1.4, 0.0 ] [x1] = [6]
	// [ 0.0, 0.2, 1.0 ] [x2]   [8]

	// x0 = 0/1 + 6/0 + 8/0 = 0
	// x2 = 0/0 + 6/0 + 8/1 = 8
	// x1 = (b1 - (x2*0.2 + x0*0.8))/1.4

	// check out http://webhome.cs.uvic.ca/~wendym/courses/445/06/5notes/2.html
	//
	//
	// TODO avoid array bounds checking where possible
	b := make(Vec, len(a))
	copy(b, a)
	if len(es) <= 1 {
		return b
	}

	// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	// fmt.Printf("Solving vec %.2f\n", b)

	xs := make(Vec, len(b))
	// for iter, e := range es {
	for iter := len(es) - 1; iter > 0; iter-- {
		e := es[iter]
		// fmt.Printf("Have E: %+v\n", e)

		if len(es) > 0 { // arbitrary?
			// call SolveDual
			var acc float32
			for j := range xs {
				if j != e.j {
					xs[j] = b[j]
					// accumulate sum against e.j
					// fmt.Printf("acc += xs[j=%v](%.2f) * e.d[j](%.2f) = %.2f\n", j, xs[j], e.d[j], xs[j]*e.d[j])
					acc += xs[j] * e.d[j]
				}
			}
			// fmt.Printf("xs[e.j=%v] = (b[e.j]{%.2f} - acc{%.2f}) / e.d[e.j]{%.2f}\n", e.j, b[e.j], acc, e.d[e.j])
			xs[e.j] = (b[e.j] - acc) / e.d[e.j]
			// TODO work out a little better
			if e.d[e.j] == 0 {
				xs[e.j] = 0
			}
		} else {
			xs[e.j] = b[e.j] / e.d[e.j]
			for j := range xs {
				if j != e.j {
					xs[j] = b[j] - e.d[j]*xs[e.j]
				}
			}
		}
		// copy x into a for next iteration
		// TODO would think it's necessary but is it?
		copy(b, xs)
		// fmt.Printf("iter %v result: %.2f\n", iter, xs)
	}
	// fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	return xs
}

func (es Etas) SolveDual(a Vec) Vec {
	// TODO avoid array bounds checking where possible
	b := make(Vec, len(a))
	copy(b, a)
	if len(es) <= 1 {
		return b
	}

	xs := make(Vec, len(b))
	for iter := len(es) - 2; iter > 0; iter-- {
		e := es[iter]
		xs[e.j] = b[e.j] / e.d[e.j]
		for j := range xs {
			if j != e.j {
				xs[j] = b[j] - e.d[j]*xs[e.j]
			}
		}
		copy(b, xs)
	}

	e := es[len(es)-1]
	var acc float32
	for j := range xs {
		if j != e.j {
			xs[j] = b[j]
			// accumulate sum against e.j
			// fmt.Printf("acc += xs[j=%v](%.2f) * e.d[j](%.2f) = %.2f\n", j, xs[j], e.d[j], xs[j]*e.d[j])
			acc += xs[j] * e.d[j]
		}
	}
	// fmt.Printf("xs[e.j=%v] = (b[e.j]{%.2f} - acc{%.2f}) / e.d[e.j]{%.2f}\n", e.j, b[e.j], acc, e.d[e.j])
	xs[e.j] = (b[e.j] - acc) / e.d[e.j]
	// TODO work out a little better
	if e.d[e.j] == 0 {
		xs[e.j] = 0
	}
	// fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	return xs
}

func (es Etas) SolveP(a Vec) Vec {
	// TODO avoid array bounds checking where possible
	b := make(Vec, len(a))
	copy(b, a)
	if len(es) <= 1 {
		return b
	}

	log.Println("b:", b)

	xs := make(Vec, len(b))
	for iter := len(es) - 1; iter > 0; iter-- {
		e := es[iter]
		log.Println("e:", e)

		// the easy way
		xs[e.j] = b[e.j] / e.d[e.j]
		for j := range xs {
			if j != e.j {
				xs[j] = b[j] - e.d[j]*xs[e.j]
			}
		}
		copy(b, xs)
	}

	return xs
}

func (es Etas) SolveFoo(a Vec) Vec {
	// TODO avoid array bounds checking where possible
	b := make(Vec, len(a))
	copy(b, a)
	if len(es) <= 1 {
		return b
	}

	log.Println("b:", b)

	xs := make(Vec, len(b))
	for iter := len(es) - 1; iter > 0; iter-- {
		e := es[iter]
		log.Println("e:", e)

		// the easy way
		xs[e.j] = b[e.j] / e.d[e.j]
		for j := range xs {
			if j != e.j {
				xs[j] = b[j] - e.d[j]*xs[e.j]
			}
		}
		copy(b, xs)
	}

	return xs
}

func greaterEqZero(xs []float32) bool {
	for _, x := range xs {
		if x < 0 {
			return false
		}
	}
	return true
}
