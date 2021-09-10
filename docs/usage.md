# Usage

The operator is looking for a *AwsAuthSyncConfig* resource in kube-system namespace to read its
configuration.

Example resource can be found in [github repo](https://raw.githubusercontent.com/gp42/aws-auth-operator/main/config/samples/auth_v1alpha1_awsauthsyncconfig.yaml).

At the moment only synchronization of IAM user groups to Kubernetes RBAC groups is supported.
For example the following configuration will look for users existing in *'source'* IAM group named
*'dev-operator-k8s-admins'* and create mappings in *'aws-auth'* ConfigMap.
```yaml
apiVersion: auth.ops42.org/v1alpha1
kind: AwsAuthSyncConfig
metadata:
  name: default
  namespace: kube-system
spec:
  syncIamGroups:
    - source: dev-operator-k8s-admins
      dest: dev-operator-k8s-admins
    - source: dev-operator-k8s-users
      dest: dev-operator-k8s-users
```

Assuming the user named *'john'* is a member of both, *'dev-operator-k8s-admins'* and *'dev-operator-k8s-users'*
groups, while user *'fred'* is only a member of the *'dev-operator-k8s-users'* group in IAM, aws-auth
ConfigMap will be modified accordingly:

```yaml
  ...
  mapUsers: |
    - userarn: arn:aws:iam::677983237296:user/john
      username: john
      groups:
      - dev-operator-k8s-admins
      - dev-operator-k8s-users
    - userarn: arn:aws:iam::677983237296:user/fred
      username: fred
      groups:
      - dev-operator-k8s-users
```

**IMPORTANT**
The operator rewrites the *data.mapUsers* part of the aws-auth configmap. Other parts remain
untouched.
