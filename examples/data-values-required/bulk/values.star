load("@ytt:data", "data")
load("@ytt:assert", "assert")

data.values.version or assert.fail("missing version")
data.values.port > 0 or assert.fail("port must be > 0")

# export
values = data.values
