package example

import (
	"log"
	"math"
	"testing"

	"github.com/mercari/testdeck"
	"github.com/stretchr/testify/assert"
)

/*
demo_test.go: This is an example of how to write Testdeck test cases
*/

// Inline Style with all four stages
func TestInlineStyle(t *testing.T) {
	// define shared variables
	var inputVariable int
	var outputVariable int

	testdeck.Test(t, &testdeck.TestCase{
		Arrange: func(t *testdeck.TD) {
			// initialize and setup shared variable values
		},
		Act: func(t *testdeck.TD) {
			// perform test code
		},
		Assert: func(t *testdeck.TD) {
			// perform output assertions here
			if inputVariable != outputVariable {
				t.Error("it's wrong!")
			}
		},
		After: func(t *testdeck.TD) {
			// perform constants cleanup here if required
		},
	})
}

// Inline style with only one stage
func TestInlineStyleMinimum(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			x := 3.0
			want := 9.0

			got := math.Pow(x, 2.0)

			if want != got {
				t.Errorf("want: %f, got %f", want, got)
			}
		},
	})
}

func TestInlineStyle_MathPow(t *testing.T) {
	var x, want, got float64

	testdeck.Test(t, &testdeck.TestCase{
		Arrange: func(t *testdeck.TD) {
			x = 3.0
			want = 9.0
		},
		Act: func(t *testdeck.TD) {
			got = math.Pow(x, 2.0)
		},
		Assert: func(t *testdeck.TD) {
			if want != got {
				t.Errorf("want: %f, got %f", want, got)
			}

		},
	})
}

// Struct style
func TestStructStyle_MathPow(t *testing.T) {
	var x, want, got float64

	test := testdeck.TestCase{}
	test.Arrange = func(t *testdeck.TD) {
		x = 3.0
		want = 9.0
	}
	test.Act = func(t *testdeck.TD) {
		got = math.Pow(x, 2.0)
	}
	test.Assert = func(t *testdeck.TD) {
		if want != got {
			t.Errorf("want: %d, got: %d", want, got)
		}
	}

	testdeck.Test(t, &test)
}

// Helper method for the test case below
func Setup(shared *int) func(t *testdeck.TD) {
	return func(t *testdeck.TD) {
		*shared = 42
	}
}

// Inline style using a helper method to Arrange that can be reused for multiple test cases
func TestReusableArrangeInline(t *testing.T) {
	var value int

	testdeck.Test(t, &testdeck.TestCase{
		Arrange: Setup(&value),
		Assert: func(t *testdeck.TD) {
			if value != 42 {
				t.Errorf("this is not the meaning of life: %d", value)
			}
			t.Logf("The meaning of life! %d", value)
		},
	})
}

// Table test style
func TestTable(t *testing.T) {
	cases := map[string]struct {
		I      int
		Result int
	}{
		"1": {
			I:      1,
			Result: 2,
		},
		"2": {
			I:      2,
			Result: 4,
		},
		"-3": {
			I:      -3,
			Result: -6,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got int
			testdeck.Test(t, &testdeck.TestCase{
				Act: func(t *testdeck.TD) {
					got = tc.I * 2
				},
				Assert: func(t *testdeck.TD) {
					if tc.Result != got {
						t.Errorf("want: %d, got: %d", tc.Result, got)
					}
				},
			})
		})
	}
}

func TestTableSequential(t *testing.T) {
	cases := map[string]struct {
		I      int
		Result int
	}{
		"1": {
			I:      1,
			Result: 2,
		},
		"2": {
			I:      2,
			Result: 4,
		},
		"-3": {
			I:      -3,
			Result: -6,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			var got int
			testdeck.Test(t, &testdeck.TestCase{
				Act: func(t *testdeck.TD) {
					log.Printf("I=%d", tc.I)
					got = tc.I * 2
				},
				Assert: func(t *testdeck.TD) {
					if tc.Result != got {
						t.Errorf("want: %d, got: %d", tc.Result, got)
					}
				},
			}, testdeck.TestConfig{ParallelOff: true})
		})
	}
}

// Table test style using shared variables
func TestSomethingSharedTable(t *testing.T) {
	cases := map[string]struct {
		In  int
		Out int
	}{
		"1": {
			In:  1,
			Out: 2,
		},
		"2": {
			In:  2,
			Out: 4,
		},
		"3": {
			In:  3,
			Out: 6,
		},
	}

	for name, tc := range cases {
		test := struct {
			val int
			testdeck.TestCase
		}{}

		test.Arrange = func(t *testdeck.TD) {
			test.val = tc.In
		}

		test.Act = func(t *testdeck.TD) {
			test.val = test.val * 2
		}

		test.Assert = func(t *testdeck.TD) {
			assert.Equal(t, tc.Out, test.val)
			// standard Go also ok:
			// if tc.Out != test.val {
			// 	t.Errorf("want: %d, got: %d", tc.Out, test.val)
			// }
		}

		test.Run(t, name)
		// test.Run is the same as:
		// t.Run(name, func(t *testing.T) {
		// 	Test(t, &test)
		// })
	}
}
