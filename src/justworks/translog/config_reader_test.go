package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_ReadSingleInput(t *testing.T) {
  config_string := "[input.foo]\nvar1=val1\nvar2=val2"

  cfg := ReadConfigurationFromString(config_string)

  assert.Equal(t, 1, len(cfg.Inputs))

  plugin := cfg.Inputs[0]

  assert.Equal(t, "foo", plugin.Name)
  assert.Equal(t, "input", plugin.T)

  assert.Equal(t, 2, len(plugin.Config))
  assert.Equal(t, "val1", plugin.Config["var1"])
  assert.Equal(t, "val2", plugin.Config["var2"])
}

func Test_ReadMultipleInputs(t *testing.T) {
  config_string := "[input.foo]\nvar1=val1\nvar2=val2\n[input.bla]\nfoo=bar"

  cfg := ReadConfigurationFromString(config_string)

  assert.Equal(t, 2, len(cfg.Inputs))

  // we already tested first plugin in test above, so just check second plugin

  plugin := cfg.Inputs[1]

  assert.Equal(t, "bla", plugin.Name)
  assert.Equal(t, "input", plugin.T)

  assert.Equal(t, 1, len(plugin.Config))
  assert.Equal(t, "bar", plugin.Config["foo"])
}

func Test_AllowKeysWithDot(t *testing.T) {
  config_string := "[input.foo]\nvar.bar=val1\nvar.1=val2\nvar3=xx"

  cfg := ReadConfigurationFromString(config_string)

  assert.Equal(t, 1, len(cfg.Inputs))

  plugin := cfg.Inputs[0]

  assert.Equal(t, "foo", plugin.Name)
  assert.Equal(t, "input", plugin.T)

  assert.Equal(t, 3, len(plugin.Config))
  assert.Equal(t, "val1", plugin.Config["var.bar"])
  assert.Equal(t, "val2", plugin.Config["var.1"])
  assert.Equal(t, "xx", plugin.Config["var3"])
}

func Test_IgnoreComments(t *testing.T) {
  config_string := "#bla\n[input.foo]\nvar1=val1\nvar2=val2\n#bla"

  cfg := ReadConfigurationFromString(config_string)

  assert.Equal(t, 1, len(cfg.Inputs))
}
