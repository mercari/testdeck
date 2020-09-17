package deferrer_test

import (
	"testing"

	"github.com/mercari/testdeck/deferrer"
	"github.com/stretchr/testify/assert"
)

func Test_DefaultDeferrer_Once(t *testing.T) {
	// Arrange
	ran := false
	fn := func() {
		ran = true
	}
	d := deferrer.DefaultDeferrer{}
	d.Defer(fn)

	// Act
	d.RunDeferred()

	// Assert
	assert.Equal(t, true, ran)
}

func Test_DefaultDeferrer_None(t *testing.T) {
	// Arrange
	d := deferrer.DefaultDeferrer{}

	// Act
	d.RunDeferred()

	// Assert
}

func Test_DefaultDeferrer_ShouldRunThreeTimes(t *testing.T) {
	// Arrange
	count := 0
	times := 3
	fn := func() {
		count++
	}
	d := deferrer.DefaultDeferrer{}
	for i := 0; i < times; i++ {
		d.Defer(fn)
	}

	// Act
	d.RunDeferred()

	// Assert
	assert.Equal(t, times, count)
}

func Test_DefaultDeferrer_ShouldExecuteInReverseOrder(t *testing.T) {
	// Arrange
	order := []int{}
	d := deferrer.DefaultDeferrer{}
	d.Defer(func() {
		order = append(order, 1)
	})
	d.Defer(func() {
		order = append(order, 2)
	})
	d.Defer(func() {
		order = append(order, 3)
	})

	// Act
	d.RunDeferred()

	// Assert
	assert.Equal(t, []int{3, 2, 1}, order)
}
