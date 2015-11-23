package simplex

import (
	"fmt"
	"testing"
)

func TestTranspose(t *testing.T) {
	a := Mat{{1, 2}, {3, 4}, {5, 6}}
	var b Mat
	a.Transpose(&b)

	if have, want := len(b), len(a[0]); have != want {
		t.Fatalf("unexpected row length, have %v, want %v", have, want)
	}

	if have, want := len(b[0]), len(a); have != want {
		t.Fatalf("unexpected column length, have %v, want %v", have, want)
	}

	want := Mat{{1, 3, 5}, {2, 4, 6}}
	for i, row := range b {
		for j, e := range row {
			if e != want[i][j] {
				t.Errorf("index %v: have %v, want %v", i, e, want[i])
			}
		}
	}
	t.Logf("\noriginal:\n%s\ntransposed:\n%s", a, b)
}

func TestBootstrap(t *testing.T) {
	prg := new(Program)
	x, y, z := prg.Var(-2), prg.Var(-3), prg.Var(-4)
	prg.AddConstraints(
		Constrain(Coef{3, x}, Coef{2, y}, Coef{1, z}).LessEq(10),
		Constrain(Coef{2, x}, Coef{5, y}, Coef{3, z}).LessEq(15),
	)
	prg.Minimize()
}

func TestTwoLines(t *testing.T) {
	// Wolfram Alpha
	// Minimize[{a + b + c + d,
	// b - a >= 50 &&
	// b - a <= 200 &&
	// c - b == 10 &&
	// d - c >= 50 &&
	// d - c <= 200 &&
	// b <= 640 && d <= 640 &&
	// a >= 0 && c >= 0},
	// {a, b, c, d}]
	prg := new(Program)
	a, b, c, d := prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	prg.AddConstraints(
		Constrain(Coef{1, b}, Coef{-1, a}).GreaterEq(50),
		Constrain(Coef{1, b}, Coef{-1, a}).LessEq(200),
		Constrain(Coef{1, c}, Coef{-1, b}).Equal(10),
		Constrain(Coef{1, d}, Coef{-1, c}).GreaterEq(50),
		Constrain(Coef{1, d}, Coef{-1, c}).LessEq(200),
		Constrain(Coef{1, b}).LessEq(640),
		Constrain(Coef{1, d}).LessEq(640),
		Constrain(Coef{1, a}).GreaterEq(0),
		Constrain(Coef{1, c}).GreaterEq(0),
	)
	prg.Minimize()
	prg.For(&a, &b, &c, &d)
	fmt.Printf("[a=%.2f][b=%.2f][c=%.2f][d=%.2f]\n", a.Val, b.Val, c.Val, d.Val)
}

func TestCircular(t *testing.T) {
	prg := new(Program)
	a0, a1, a2, a3 := prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	b0, b1, b2, b3 := prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	c0, c1, c2, c3 := prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	prg.AddConstraints(
		Constrain(Coef{1, a1}, Coef{-1, a0}).Equal(200),
		Constrain(Coef{1, a3}, Coef{-1, a2}).Equal(200),
		Constrain(Coef{1, b0}, Coef{-1, a1}).GreaterEq(0),
		Constrain(Coef{1, c0}, Coef{-1, a1}).GreaterEq(0),
		Constrain(Coef{1, a0}).GreaterEq(0),
		Constrain(Coef{1, a1}).LessEq(640),
		Constrain(Coef{1, a2}).GreaterEq(0),
		Constrain(Coef{1, a3}).LessEq(480),
		Constrain(Coef{1, b1}, Coef{-1, b0}).Equal(50),
		Constrain(Coef{1, b3}, Coef{-1, b2}).Equal(50),
		Constrain(Coef{1, c0}, Coef{-1, b1}).GreaterEq(0),
		Constrain(Coef{1, b0}).GreaterEq(0),
		Constrain(Coef{1, b1}).LessEq(640),
		Constrain(Coef{1, b2}).GreaterEq(0),
		Constrain(Coef{1, b3}).LessEq(480),
		Constrain(Coef{1, c1}, Coef{-1, c0}).Equal(100),
		Constrain(Coef{1, c3}, Coef{-1, c2}).Equal(400),
		Constrain(Coef{1, c0}).GreaterEq(0),
		Constrain(Coef{1, c1}).LessEq(640),
		Constrain(Coef{1, c2}).GreaterEq(0),
		Constrain(Coef{1, c3}).LessEq(480),
	)
	prg.Minimize()
	prg.For(&a0, &a1, &a2, &a3, &b0, &b1, &b2, &b3, &c0, &c1, &c2, &c3)
	fmt.Printf("[a0=%.2f][a1=%.2f][a2=%.2f][a3=%.2f]\n", a0.Val, a1.Val, a2.Val, a3.Val)
	fmt.Printf("[b0=%.2f][b1=%.2f][b2=%.2f][b3=%.2f]\n", b0.Val, b1.Val, b2.Val, b3.Val)
	fmt.Printf("[c0=%.2f][c1=%.2f][c2=%.2f][c3=%.2f]\n", c0.Val, c1.Val, c2.Val, c3.Val)
}
