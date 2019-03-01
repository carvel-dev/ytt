def labelSafe(str):
  return str.replace("+", "_")
end

def labelSafe2(str):
  return str[:63].rstrip("-")
end
