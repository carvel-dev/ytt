load("@ytt:data", "data")

# Version of prometheus
version = "v0.39.0"

deployment = {
  "name": data.values.app_name,
  "containerPort": 8080
}