package simplex

import "testing"

func _TestRevisedMaximize(t *testing.T) {
	// x1=2, x2=3, z=27
	prg := new(Revised)
	prg.c = Vec{6, 5, 0, 0}
	prg.a = Mat{
		{1, 1, 1, 0},
		{3, 2, 0, 1},
	}
	prg.b = Vec{5, 12}
	prg.Maximize()
}

func _TestRevisedWolfram(t *testing.T) {
	// Maximize[{a + 2b, a + b <= 10 && 2a + 3b <= 25 && a + 5b <= 35}, {a, b}]
	// Z=110/7 a=20/7 b=45/7
	prg := new(Revised)
	prg.c = Vec{1, 2, 0, 0}
	prg.a = Mat{
		{2, 3, 1, 0},
		{1, 5, 0, 1},
	}
	prg.b = Vec{25, 35}
	prg.Maximize()
}

func _TestRevisedBootstrap(t *testing.T) {
	// prg := new(Revised)
	// prg.c = Vec{-2, -3, -4, 0, 0}
	// prg.a = Mat{
	// {3, 2, 1, 1, 0},
	// {2, 5, 3, 0, 1},
	// }
	// prg.b = Vec{10, 15}
	// prg.Minimize()

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
	// prg := new(Revised)
	// prg.c = Vec{
	// -1, -1, -1, -1, 0, 0, 0, 0, 0,
	// }
	// prg.a = Mat{ //..
	// {-1, 1, 0, 0, -1, 0, 0, 0, 0},
	// {0, -1, 1, 0, 0, 1, 0, 0, 0},
	// {0, 0, -1, 1, 0, 0, -1, 0, 0},
	// {0, 1, 0, 0, 0, 0, 0, 1, 0},
	// {0, 0, 0, 1, 0, 0, 0, 0, 1},
	// }
	// prg.b = Vec{50, 10, 50, 640, 640}
	// prg.Minimize()

	// https://www.youtube.com/watch?v=eGqA94MCV9c
	// prg := new(Revised)
	// prg.c = Vec{1, -1, 1, -1, 0, 0, 0, 0}
	// prg.a = Mat{
	// {2, -3, 7, -15, 1, 0, 0, 0},
	// {0, 1, -4, 6, 0, 1, 0, 0},
	// {-1, 0, 1, -2, 0, 0, 1, 0},
	// {0, 1, 1, 0, 0, 0, 0, 1},
	// }
	// prg.b = Vec{10, 12, 4, 16}
	// prg.Maximize3()

	// https://www.youtube.com/watch?v=UsqtzA9XmQE
	// Maximize[{6a + 8b, a + b <= 10 && 2a + 3b <= 25 && a + 5b <= 35}, {a, b}]
	// Z=70 a=5 b=5
	prg := new(Revised)
	prg.c = Vec{6, 8, 0, 0, 0}
	prg.a = Mat{
		{1, 1, 1, 0, 0},
		{2, 3, 0, 1, 0},
		{1, 5, 0, 0, 1},
	}
	prg.b = Vec{10, 25, 35}
	prg.Maximize()

	// Maximize[{a + 2b, a + b <= 10 && 2a + 3b <= 25 && a + 5b <= 35}, {a, b}]
	// Z=110/7 a=20/7 b=45/7
	// prg := new(Revised)
	// prg.c = Vec{1, 2, 0, 0}
	// prg.a = Mat{
	// {2, 3, 1, 0},
	// {1, 5, 0, 1},
	// }
	// prg.b = Vec{25, 35}
	// prg.Maximize4()
}

// TODO this is given as minimize on wikipedia, just write out dual to use for test for maximize
// func TestBootstrap3(t *testing.T) {
// Maximize[{2a + 3b + 4c, 3a + 2b + c <= 10 && 2a + 5b + 3c <= 15 && a >= 0 && b >= 0 & c >= 0}, {a, b, c}]
// no global maxima without Xi >= 0
// prg := new(Revised)
// prg.c = Vec{2, 3, 4, 0, 0}
// prg.a = Mat{
// {3, 2, 1, 1, 0},
// {2, 5, 3, 0, 1},
// }
// prg.b = Vec{10, 15}
// prg.Maximize()
// }

func _TestRevisedTwoLines(t *testing.T) {
	// Wolfram Alpha
	// Maximize[{a + b + c + d, b - a <= 50 && c - b <= 10 && d - c <= 50 && b <= 640 && d <= 640 && a >= 0 && c >= 0}, {a, b, c, d}]
	prg := new(Revised)
	prg.c = Vec{1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0}
	prg.a = Mat{
		{-1, 1, 0, 0, -1, 0, 0, 0, 0, 0, 0},
		{0, -1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, -1, 1, 0, 0, -1, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, -1, 0},
		{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, -1},
	}
	prg.b = Vec{50, 10, 50, 640, 640, 0, 0}
	prg.Maximize()
}
