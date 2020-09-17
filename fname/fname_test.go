package fname

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Fname_ShouldReturnFuncName(t *testing.T) {
	// Arrange

	// Act
	fn := Fname()

	// Assert
	assert.Equal(t, t.Name(), fn)
}

func Test_Fname_ShouldReturnParentLevelFuncName(t *testing.T) {
	// Arrange

	// Act
	fn := func() string {
		return Fname(1)
	}()

	// Assert
	assert.Equal(t, t.Name(), fn)
}

func Test_Fname_ShouldReturnGrandParentLevelFuncName(t *testing.T) {
	// Arrange

	// Act
	fn := func() string {
		return func() string {
			return Fname(2)
		}()
	}()

	// Assert
	assert.Equal(t, t.Name(), fn)
}
