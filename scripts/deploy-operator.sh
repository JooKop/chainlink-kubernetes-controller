# This script builds the chainlink-kubernetes-operator project into a Docker image
# The created image is then pushed into AWS ECR
export IMAGE_TAG_BASE="310341869582.dkr.ecr.eu-central-1.amazonaws.com/chainlink-kubernetes-operator"
(cd operator && make deploy)