module github.com/gp42/aws-auth-operator

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.7.1
	github.com/aws/aws-sdk-go-v2/config v1.5.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.7.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
