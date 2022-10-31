package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_openCSV(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	file, err := openCSV("../test/test..csv")
	assert.NotNil(t, err)
	assert.Nil(t, file)

	file, err = openCSV("../test/test.csv")
	assert.Nil(t, err)
	assert.NotNil(t, file)
}
