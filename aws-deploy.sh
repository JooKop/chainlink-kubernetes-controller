TEMPLATE_FILE="aws/eks-iam-cluster-vpc-subnets.yaml"
STACK_NAME="chainlink-hackathon"

# Parameters
IAM_ROLE_NAME="EKSAdmin"

echo "Deleting stack '$STACK_NAME' if exists..."
aws cloudformation delete-stack --stack-name "$STACK_NAME"
aws cloudformation wait stack-delete-complete --stack-name "$STACK_NAME"
echo "Deploying stack '$STACK_NAME'..."
aws cloudformation deploy \
--template-file "$TEMPLATE_FILE" \
--stack-name "$STACK_NAME" \
--parameter-overrides IAMRoleName="$IAM_ROLE_NAME" \
--capabilities CAPABILITY_NAMED_IAM