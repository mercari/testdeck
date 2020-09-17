package example

import (
	"testing"

	"github.com/mercari/testdeck"
)

/*
err_handling_test.go: Examples of how to use error handling and assertions in test cases
*/

func Test_BasicExample_ActError(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Errorf("basic act example error")
		},
	})
}

func Test_BasicExample_ActFatal(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Fatalf("basic act example fatal")
		},
	})
}

func Test_BasicExample_AssertErrorAndFatal(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Assert: func(t *testdeck.TD) {
			t.Errorf("basic assert example error")
			t.Fatalf("basic assert example fatal")
		},
	})
}

func Test_BasicExample_ArrangeMultiFatal(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Arrange: func(t *testdeck.TD) {
			t.Fatalf("basic arrange example fatal 1")
			t.Fatalf("basic arrange example fatal 2")
		},
	})
}

func Test_BasicExample_ArrangeAfterMultiFatal(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Arrange: func(t *testdeck.TD) {
			t.Fatalf("basic arrange example fatal 1")
		},
		After: func(t *testdeck.TD) {
			t.Fatalf("basic after example fatal 2")
		},
	})
}

func Test_BasicExample_ArrangeFatalAndDeferred(t *testing.T) {
	testCase := &testdeck.TestCase{}
	testCase.Arrange = func(t *testdeck.TD) {
		testCase.Defer(func() {
			t.Errorf("basic deferred output")
		})
		t.Fatalf("basic arrange example fatal 1")
		t.Fatalf("basic arrange example fatal 2")
	}

	testdeck.Test(t, testCase)
}

func Test_ActEmptyNoError(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {},
	})
}

func Test_ArrangeSkipNow_ShouldbeMarkedSkip(t *testing.T) {
	test := &testdeck.TestCase{}
	test.Arrange = func(t *testdeck.TD) {
		test.Defer(func() {
			t.Log("skip now defer message 1")
		})
		t.SkipNow()
		test.Defer(func() {
			t.Log("skip now defer message 2")
		})
	}
	test.Act = func(t *testdeck.TD) {
		t.Error("skip now act error")
	}
	test.Assert = func(t *testdeck.TD) {
		t.Error("skip now assert error")
	}
	test.After = func(t *testdeck.TD) {
		t.Log("skip now after message")
	}
	testdeck.Test(t, test)
}

func Test_ActSkipNow_ShouldbeMarkedSkipAndExecuteAfter(t *testing.T) {
	test := &testdeck.TestCase{}
	test.Arrange = func(t *testdeck.TD) {
		test.Defer(func() {
			t.Log("skip now defer message")
		})
	}
	test.Act = func(t *testdeck.TD) {
		t.SkipNow()
		t.Error("skip now act error")
	}
	test.Assert = func(t *testdeck.TD) {
		t.Error("skip now assert error")
	}
	test.After = func(t *testdeck.TD) {
		t.Log("skip now after message")
	}
	testdeck.Test(t, test)
}
