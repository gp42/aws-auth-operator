# Installation

Installation Options

- [Operator Lifecycle Manager (OLM)](#operator-lifecycle-manager-olm)
- [Manual Installation](#manual-installation)

## Operator Lifecycle Manager (OLM)
Check the [deploy](https://github.com/gp42/aws-auth-operator/tree/main/deploy) directory for manifest examples.
These instructions assume that you have OLM installed in the default **olm** namespace.

**Install catalog source**

```bash
kubectl apply -n olm -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/olm/catalogsource.yaml
# verify
kubectl describe catalogsource -n olm aws-auth-operator-catalog
```

**Create an OperatorGroup**
```bash
kubectl create namespace aws-auth-operator-system
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/olm/operatorgroup.yaml
```

**Install subscription**

Make sure to set AWS secrets for the operator.
See [AWS User Policy](#aws-user-policy) section for required access for this user.

This example shows how to set the secrets using 'Secret' resource:

```bash
kubectl create secret generic \
  -n aws-auth-operator-system \
  aws-auth-operator-secret \
    --from-literal=AWS_ACCESS_KEY_ID="<key>" \
    --from-literal=AWS_SECRET_ACCESS_KEY="<secret>" \
    --from-literal=AWS_DEFAULT_REGION="<region>"

kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/olm/subscription.yaml
```

**Approve InstallPlan**

Wait for subsctiption, you can check current status with the following commands:
```bash
kubectl get subscriptions -n aws-auth-operator-system aws-auth-operator
kubectl describe subscription -n aws-auth-operator-system aws-auth-operator
```

Manually approve the InstallPlan:
```bash
kubectl get installplans -n aws-auth-operator-system
kubectl patch installplan <InstallPlan Name> --type merge --patch '{"spec": {"approved": true}}'
```

If the InstallPlan does not appear, check olm logs:
```bash
kubectl logs -f -n olm <olm-operator-xxx pod name>
```

Check if the operator was successfully deployed:
```bash
kubectl get csv -n aws-auth-operator-system
kubectl get pods -n aws-auth-operator-system
```

## Manual Installation
**Namespace**

Create a new namespace for the operator
```bash
kubectl create namespace aws-auth-operator-system
```

**Install CRDs**

Install Custom Resource Definitions
```bash
kubectl apply -n aws-auth-operator-system -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/crds.yaml
```

**Create Secrets**

Make sure to set AWS secrets for the operator.
See [AWS User Policy](#aws-user-policy) section for required access for this user.

This example shows how to set the secrets using 'Secret' resource:

```bash
kubectl create secret generic \
  -n aws-auth-operator-system \
  aws-auth-operator-secret \
    --from-literal=AWS_ACCESS_KEY_ID="<key>" \
    --from-literal=AWS_SECRET_ACCESS_KEY="<secret>" \
    --from-literal=AWS_DEFAULT_REGION="<region>"
```

**Install all other resources**
```bash
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/serviceaccount.yaml
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/deployment.yaml
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/role.yaml
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/role_binding.yaml
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/role_leader_election.yaml
kubectl apply -f https://raw.githubusercontent.com/gp42/aws-auth-operator/main/deploy/manual/role_binding_leader_election.yaml
```

** 

## AWS User Policy

AWS user which keys are provided to the operator, must have the following policy attached to be able
to do IAM group scanning:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "iam:GetGroup",
            "Resource": "*"
        }
    ]
}
```

## Usage
See [Usage](usage.md) section.
