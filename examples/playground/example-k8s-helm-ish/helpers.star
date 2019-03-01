def name(vars):
  return kube_clean_name(vars.Chart.Name or vars.Values.nameOverride)
end

def fullname(vars):
  name = vars.Chart.Name or vars.Values.nameOverride
  return kube_clean_name("{}-{}".format(vars.Release.Name, name))
end

def kube_clean_name(name):
  return name[:63].rstrip("-")
end
