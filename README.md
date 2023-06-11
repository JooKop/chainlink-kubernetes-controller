## 🎉 This is a Chainlink Spring Hackathon 2023 project 🎉

# chainlink-kubernetes-operator
This project implements a Kubernetes Operator that simplifies the running of Chainlink oracles and their jobs in Kubernetes clusters via the use of Custom Resource Definitions (CRDs).

The repository contains four different folders with the following purposes:
1. [aws](/aws/) - CloudFormation templates for deploying the following AWS resources: Elastic Kubernetes Service, EC2 worker nodes and related IAM roles.
2. [examples](/examples/) - Once you have the `chainlink-kubernetes-operator` running in a Kubernetes cluster, you can use these custom resources to deploy nodes and their jobs.
3. [operator](/operator/) - This is the main `chainlink-kubernetes-operator` project folder. It contains an [operator-sdk](https://sdk.operatorframework.io/) bootstrapped project
4. [scripts](/scripts/) - These scripts can be used to deploy the AWS EKS cluster from scratch, build and push the `chainlink-kubernetes-operator` image to a container registry and to deploy the operator to a Kubernetes cluster (whichever is the current context in `~/.kube/config`).

Everything was built during the [Chainlink Spring Hackathon 2023](https://chain.link/hackathon) and fully operational.