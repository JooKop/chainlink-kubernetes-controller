# This script builds the chainlink-kubernetes-operator project into a Docker image
# The created image is then pushed into AWS ECR
ECR_REGISTRY="310341869582.dkr.ecr.eu-central-1.amazonaws.com/chainlink-kubernetes-operator"
REGION="eu-central-1"
IMAGE_TAG_BASE="$ECR_REGISTRY/chainlink-kubernetes-controller"

aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

(cd operator && make docker-build docker-push)