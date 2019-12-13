load("@ytt:data", "data")
load("@ytt:assert", "assert")

data.values.name or assert.fail("missing name")
data.values.namespace or assert.fail("missing namespace")
data.values.image or assert.fail("missing image")
data.values.port > 0 or assert.fail("port expected to be > 0")
