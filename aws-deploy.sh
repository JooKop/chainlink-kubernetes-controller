TEMPLATE_FILE="aws/deployment.yaml"
STACK_NAME="hackathon4"

# Parameters
IAM_ROLE_NAME="EKSAdmin"

aws cloudformation destroy-stack --stack-name $STACK_NAME
aws cloudformation deploy \
--template-file "$TEMPLATE_FILE" \
--stack-name "$STACK_NAME" \
--parameter-overrides EKSIAMRoleName="$IAM_ROLE_NAME" \
--capabilities CAPABILITY_NAMED_IAM