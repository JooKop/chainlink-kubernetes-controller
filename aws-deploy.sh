aws cloudformation deploy \
--template-file aws/deployment.yaml \
--stack-name hackathon4 \
--parameter-overrides EKSIAMRoleName=eksadmin \
--capabilities CAPABILITY_NAMED_IAM