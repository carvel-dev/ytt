load("@ytt:data", "data")

# Version of prometheus
version = "v0.39.0"

# Shared values from Service
service = {
  "name": "{}-service".format(data.values.app_name),
  "port": 38080
}

deployment = {
  "name": data.values.app_name,
  "containerPort": 8080
}