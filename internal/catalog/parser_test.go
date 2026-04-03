package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
)

func TestParse_ValidYAML(t *testing.T) {
	manifest := `                                                                                                                                                                                                                                       
name: webserver

resources:
  - package:
      nginx:
        state: present
`

	c, err := Parse([]byte(manifest))
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "webserver", c.Name)
	assert.Equal(t, 1, len(c.Resources))
	assert.Equal(t, "nginx", c.Resources[0].Name())
	assert.Equal(t, "package", c.Resources[0].Type())
}

func TestParse_InvalidYAML(t *testing.T) {
	manifest := `
name: broken
resources:
  - this is: [not: valid
`

	c, err := Parse([]byte(manifest))
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestParse_UnknownResourceType(t *testing.T) {
	manifest := `
name: test

resources:
  - foobar:
      something:
        state: present
`

	c, err := Parse([]byte(manifest))
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestParse_MissingRequiredAttribute(t *testing.T) {
	manifest := `
name: test

resources:
  - package:
      nginx:
        invalid: present
`

	c, err := Parse([]byte(manifest))
	assert.NotNil(t, err)
	assert.Nil(t, c)
}
