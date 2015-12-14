package simplex

import "testing"

const epsilon = 0.0001

func equals(a, b float32) bool {
	return equaleps(a, b, epsilon)
}

func equaleps(a, b float32, eps float32) bool {
	return (a-b) < eps && (b-a) < eps
}

func TestMaximize(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(6), prg.Var(5)
	prg.AddConstraints(
		Constrain(Coef{1, x1}, Coef{1, x2}).LessEq(5),
		Constrain(Coef{3, x1}, Coef{2, x2}).LessEq(12),
	)

	if err := prg.Maximize(); err != nil {
		t.Log(err)
	}

	prg.For(&x1, &x2)
	var wz, w1, w2 float32 = 27, 2, 3
	if z := prg.Z(); !equals(z, wz) || !equals(x1.Val, w1) || !equals(x2.Val, w2) {
		t.Fatalf("unexpected result:\nHave Z=%.2f x1=%.2f x2=%.2f\nWant Z=%.2f x1=%.2f x2=%.2f\n%s\n", z, x1.Val, x2.Val, wz, w1, w2, prg.tbl)
	}
}

func TestMinimize(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(3), prg.Var(4)
	prg.AddConstraints(
		Constrain(Coef{2, x1}, Coef{3, x2}).GreaterEq(8),
		Constrain(Coef{5, x1}, Coef{2, x2}).GreaterEq(12),
	)

	if err := prg.Minimize(); err != nil {
		t.Log(err)
	}

	prg.For(&x1, &x2)
	var wz, w1, w2 float32 = 124.0 / 11, 20.0 / 11, 16.0 / 11
	if z := prg.Z(); !equals(z, wz) || !equals(x1.Val, w1) || !equals(x2.Val, w2) {
		t.Fatalf("unexpected result:\nHave Z=%.2f x1=%.2f x2=%.2f\nWant Z=%.2f x1=%.2f x2=%.2f\n%s\n", z, x1.Val, x2.Val, z, w1, w2, prg.tbl)
	}
}

func TestDegeneracy(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(4), prg.Var(3)
	prg.AddConstraints(
		Constrain(Coef{2, x1}, Coef{3, x2}).LessEq(8),
		Constrain(Coef{3, x1}, Coef{2, x2}).LessEq(12),
	)

	if err := prg.Maximize(); err != nil {
		t.Log(err)
	}

	prg.For(&x1, &x2)
	var wz, w1, w2 float32 = 16, 4, 0
	if z := prg.Z(); !equals(z, wz) || !equals(x1.Val, w1) || !equals(x2.Val, w2) {
		t.Fatalf("unexpected result:\nHave Z=%.2f x1=%.2f x2=%.2f\nWant Z=%.2f x1=%.2f x2=%.2f\n%s\n", z, x1.Val, x2.Val, wz, w1, w2, prg.tbl)
	}
}

func TestAlternateOptimum(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(4), prg.Var(3)
	prg.AddConstraints(
		Constrain(Coef{8, x1}, Coef{6, x2}).LessEq(25),
		Constrain(Coef{3, x1}, Coef{4, x2}).LessEq(15),
	)

	if err := prg.Maximize(); err != nil {
		t.Log(err)
	}

	prg.For(&x1, &x2)
	var wz, w1, w2 float32 = 25.0 / 2, 5.0 / 7, 45.0 / 14
	if z := prg.Z(); !equals(z, wz) || !equals(x1.Val, w1) || !equals(x2.Val, w2) {
		t.Fatalf("unexpected result:\nHave Z=%.2f x1=%.2f x2=%.2f\nWant Z=%.2f x1=%.2f x2=%.2f\n%s\n", z, x1.Val, x2.Val, wz, w1, w2, prg.tbl)
	}
}

func TestUnbounded(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(4), prg.Var(3)
	prg.AddConstraints(
		Constrain(Coef{1, x1}, Coef{-6, x2}).LessEq(5),
		Constrain(Coef{3, x1}).LessEq(11),
	)

	if err := prg.Maximize(); err != ErrUnbounded {
		t.Fatal("expected program to terminate ErrUnbounded")
	}

	// TODO this problem has an optimal solution for minimization, break this out into another
	// test for minimization with only LessEq to uncover inconsistent states in c, a, b.
	if err := prg.Minimize(); err != nil {
		t.Fatal(err)
	}
}

func TestInfeasible(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(4), prg.Var(3)
	prg.AddConstraints(
		Constrain(Coef{1, x1}, Coef{4, x2}).LessEq(3),
		Constrain(Coef{3, x1}, Coef{1, x2}).GreaterEq(12),
	)

	if err := prg.Maximize(); err != ErrInfeasible {
		t.Fatal("expected program to terminate ErrInfeasible")
	}
}

func TestEquations(t *testing.T) {
	prg := new(Program)
	x1, x2 := prg.Var(1), prg.Var(1)
	prg.AddConstraints(
		Constrain(Coef{2, x1}, Coef{3, x2}).Equal(6),
		Constrain(Coef{4, x1}, Coef{6, x2}).Equal(12),
		Constrain(Coef{1, x1}).GreaterEq(0),
		Constrain(Coef{1, x2}).GreaterEq(0),
	)

	if err := prg.Minimize(); err != nil {
		t.Log(err)
	}

	prg.For(&x1, &x2)
	var wz, w1, w2 float32 = 2, 0, 2
	if z := prg.Z(); !equals(z, wz) || !equals(x1.Val, w1) || !equals(x2.Val, w2) {
		t.Fatalf("unexpected result:\nHave Z=%.2f x1=%.2f x2=%.2f\nWant Z=%.2f x1=%.2f x2=%.2f\n%s\n", z, x1.Val, x2.Val, wz, w1, w2, prg.tbl)
	}
}

func TestTwoLines(t *testing.T) {
	// Minimize[{a + b + c + d, b - a >= 50 && c - b == 10 && d - c >= 50 && b <= 640 && d <= 640 && a >= 0 && c >= 0}, {a, b, c, d}]
	prg := new(Program)
	a, b, c, d := prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	prg.AddConstraints(
		Constrain(Coef{1, b}, Coef{-1, a}).GreaterEq(50),
		Constrain(Coef{1, c}, Coef{-1, b}).Equal(10),
		Constrain(Coef{1, d}, Coef{-1, c}).GreaterEq(50),
		Constrain(Coef{1, b}).LessEq(640),
		Constrain(Coef{1, d}).LessEq(640),
		Constrain(Coef{1, a}).GreaterEq(0),
		Constrain(Coef{1, c}).GreaterEq(0),
	)

	if err := prg.Minimize(); err != nil {
		t.Log(err)
	}

	prg.For(&a, &b, &c, &d)
	var wz, wa, wb, wc, wd float32 = 220, 0, 50, 60, 110
	if z := prg.Z(); !equals(z, wz) || !equals(a.Val, wa) || !equals(b.Val, wb) || !equals(c.Val, wc) || !equals(d.Val, wd) {
		t.Fatalf("unexpected result:\nHave Z=%.2f a=%.2f b=%.2f c=%.2f d=%.2f\nWant Z=%.2f a=%.2f b=%.2f c=%.2f d=%.2f\n%s\n", z, a.Val, b.Val, c.Val, d.Val, wz, wa, wb, wc, wd, prg.tbl)
	}
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
	if err := prg.Minimize(); err != nil {
		t.Log(err)
	}
	prg.For(
		&a0, &a1, &a2, &a3,
		&b0, &b1, &b2, &b3,
		&c0, &c1, &c2, &c3,
	)
	t.Logf("[a0=%.2f][a1=%.2f][a2=%.2f][a3=%.2f]\n", a0.Val, a1.Val, a2.Val, a3.Val)
	t.Logf("[b0=%.2f][b1=%.2f][b2=%.2f][b3=%.2f]\n", b0.Val, b1.Val, b2.Val, b3.Val)
	t.Logf("[c0=%.2f][c1=%.2f][c2=%.2f][c3=%.2f]\n", c0.Val, c1.Val, c2.Val, c3.Val)
}

func BenchmarkMaximize(b *testing.B) {
	prg := new(Program)
	x1, x2 := prg.Var(6), prg.Var(5)
	prg.AddConstraints(
		Constrain(Coef{1, x1}, Coef{1, x2}).LessEq(5),
		Constrain(Coef{3, x1}, Coef{2, x2}).LessEq(12),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := prg.Maximize(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMinimize(b *testing.B) {
	prg := new(Program)
	x1, x2 := prg.Var(3), prg.Var(4)
	prg.AddConstraints(
		Constrain(Coef{2, x1}, Coef{3, x2}).GreaterEq(8),
		Constrain(Coef{5, x1}, Coef{2, x2}).GreaterEq(12),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := prg.Minimize(); err != nil {
			b.Fatal(err)
		}
	}
}
