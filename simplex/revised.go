package simplex

import (
	"fmt"
	"log"
)

type Revised struct {
	c Vec
	a Mat
	b Vec
}

func (prg *Revised) Maximize() {
	fmt.Printf("LinearProgram:\nc:\n%.2f\nA:\n%sb:\n%.2f\n", prg.c, prg.a, prg.b)
	fmt.Println("------------")
	var B []int // Basis; indices of basic vars in objective function.
	var I []int // Independent; indices of non-basic vars in objective function.
	for i, x := range prg.c {
		if x == 0 {
			B = append(B, i)
		} else {
			I = append(I, i)
		}
	}
	fmt.Printf("B (Basis; indices of basic vars in objective function.)\n%+v\n", B)
	fmt.Printf("I (Independent; indicies of non-basic vars in objective function.)\n%+v\n", I)

	// TODO don't need to store entire matrix, just vector and column to apply to identity
	// TODO should probably size so column is the row to ease computation
	var Es = Etas{{make(Vec, len(prg.a)), 0}} // Eta matrices
	Es[0].d[0] = 1
	fmt.Printf("E0 (Eta Matrix): %+v\n", Es[0])

	XB := make(Vec, len(prg.b))
	copy(XB, prg.b)

	// var Solution Vec

	const nsteps = 99
	for step := 0; step < nsteps; step++ {
		fmt.Printf("########## STEP %v ##########\n", step+1)
		// Check non-basic variables for if current Basis is optimal
		// Ci - Zi = Ci - y*Pi
		// Ci is objective function coefficient of Xi
		// y is the value of the dual
		// Pi is the vector corresponding to the variable Xi in the original problem

		var C Vec
		var P Mat
		for _, x := range I {
			C = append(C, prg.c[x])
			P = append(P, prg.a.ColumnAlloc(x))
		}
		fmt.Printf("C: %.2f\n", C)
		fmt.Printf("P: %.2f\n", P)

		// y = CB * B^(-1)
		// CB is objective function coefficients of basic variables
		var CB Vec
		for _, x := range B {
			CB = append(CB, prg.c[x])
		}
		fmt.Printf("CB: %.2f\n", CB)

		var y Vec = Es.SolveDual(CB)
		fmt.Printf("y: %.2f\n", y)

		var Ci Vec = make(Vec, len(C))
		C.Sub(&Ci, P.MulVec(y))
		fmt.Printf("Ci: %.2f\n", Ci)

		var jj int
		var max float32
		if jj, max = Ci.Max(); max <= 0 { // maximum coefficient rule
			log.Println("simplex: optimal solution reached")
			log.Printf("solution: %+v = %.2f\n", B, XB)
			return
		}
		fmt.Printf("I[jj=%v] (entering variable): %v\n", jj, I[jj])

		// compute leaving variable
		// Pbar = B^(-1) * Pj
		Pj := prg.a.ColumnAlloc(jj)
		Pbar := Es.SolveP(Pj)
		fmt.Printf("Pj: %.2f\n", Pj)
		fmt.Printf("Pbar: %.2f\n", Pbar)

		// fmt.Printf("SolveP for XBb: %.2f\n", Es.SolveP(prg.b))
		// XB := Es.SolveP(prg.b)
		fmt.Printf("XB: %.2f\n", XB)
		// fmt.Printf("XB.DivGZ(Pbar): %.2f\n", XB.DivGZ(Pbar))
		// TODO sort div vec with XB so Min() takes var with smallest subscript if there's multiple
		ii, theta := Theta(XB, Pbar) // XB.DivGZ(Pbar).Min()
		if theta <= 0 {
			fmt.Printf("!!! theta <= 0: %+v = %.2f\n", XB, Es.SolveP(prg.b))
			return
		}
		fmt.Printf("theta: %.2f\n", theta)
		fmt.Printf("XB[ii=%v] (leaving variable): %v\n", ii, B[ii])

		// Swap entering and leaving variables
		fmt.Println("Swap entering and leaving variables")
		fmt.Printf("Before:\nB: %v\nI: %v\n", B, I)
		B[ii], I[jj] = I[jj], B[ii]
		fmt.Printf("After:\nB: %v\nI: %v\n", B, I)

		pbarcpy := make(Vec, len(Pbar))
		copy(pbarcpy, Pbar)
		Es = append(Es, Eta{pbarcpy, ii})
		// pjcpy := make(Vec, len(Pj))
		// copy(pjcpy, Pj)
		// Es = append(Es, Eta{pjcpy, ii})
		for i, e := range Es {
			fmt.Printf("[i=%v] %+v\n", i, e)
		}

		// update solution
		Pbar.MulScalar(&Pbar, theta)
		// Solution = XB.Sub(Pbar)
		// Solution[ii] = theta
		XB.Sub(&XB, Pbar)
		XB[ii] = theta
	}
}
