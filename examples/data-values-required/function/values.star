load("@ytt:data", "data")
load("@ytt:assert", "assert")

def version():
  return data.values.version or assert.fail("missing version")
end

def port():
  if data.values.port > 0:
    return data.values.port
  end
  assert.fail("port must be > 0")
end
