## Running on AWS ALB+Lambda

- run `./hack/build.sh` which generates `./tmp/ytt-lambda-website.zip`
- create new Lambda with Go 1.x runtime and handler "main"
  - use above zip for the upload
- create new ALB and set its attribute "Multi-value header" to checked
- access https://<alb-dns>
