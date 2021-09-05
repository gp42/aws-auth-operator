package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authv1alpha1 "github.com/gp42/aws-auth-operator/api/v1alpha1"
)

var _ = Describe("CronJob controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		AwsAuthSyncConfigName = "default"
		AwsAuthCmName         = "aws-auth"
		AwsAuthCmNamespace    = "kube-system"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating AwsAuthSyncConfig Status", func() {
		It("Should increase AwsAuthSyncConfig Status", func() {
			ctx := context.Background()
			By("Creating a new aws-auth ConfigMap")
			awsAuthCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AwsAuthCmName,
					Namespace: AwsAuthCmNamespace,
				},
				Data: map[string]string{},
			}

			Expect(k8sClient.Create(ctx, awsAuthCm)).Should(Succeed())

			By("By creating a new AwsAuthSyncConfig")
			AwsAuthSyncConfig := &authv1alpha1.AwsAuthSyncConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "auth.ops42.org/v1alpha1",
					Kind:       "AwsAuthSyncConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      AwsAuthSyncConfigName,
					Namespace: AwsAuthCmNamespace,
				},
				Spec: authv1alpha1.AwsAuthSyncConfigSpec{
					SyncIamGroups: []authv1alpha1.AwsAuthSyncConfigGroupSync{{
						Source: "mercenary-admins",
						Dest:   "mercenary-admins",
					}, {
						Source: "mercenary-users",
						Dest:   "mercenary-users",
					}, {
						Source: "jedi-admins",
						Dest:   "jedi-admins",
					}},
				},
			}
			Expect(k8sClient.Create(ctx, AwsAuthSyncConfig)).Should(Succeed())

			awsAuthConfigLookupKey := client.ObjectKey{
				Name:      AwsAuthSyncConfigName,
				Namespace: AwsAuthCmNamespace,
			}
			createdAwsAuthSyncConfig := &authv1alpha1.AwsAuthSyncConfig{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, awsAuthConfigLookupKey, createdAwsAuthSyncConfig)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdAwsAuthSyncConfig.Spec.SyncIamGroups[0].Source).Should(Equal("mercenary-admins"))
			Expect(createdAwsAuthSyncConfig.Spec.SyncIamGroups[1].Source).Should(Equal("mercenary-users"))
			Expect(createdAwsAuthSyncConfig.Spec.SyncIamGroups[2].Source).Should(Equal("jedi-admins"))

			By("By updating aws-auth ConfigMap")
			synchedAwsAuthSyncConfig := &authv1alpha1.AwsAuthSyncConfig{}

			By("By checking the AwsConfig Sync status")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, awsAuthConfigLookupKey, synchedAwsAuthSyncConfig)
				if err != nil {
					return "", err
				}
				return synchedAwsAuthSyncConfig.Status.Status, nil
			}, timeout, interval).Should(Equal("Success"))

			By("By checking synched result at aws-auth ConfigMap")
			awsAuthCmLookupKey := client.ObjectKey{Name: AwsAuthCmName, Namespace: AwsAuthCmNamespace}
			synchedAwsAuthCm := &corev1.ConfigMap{}

			expectedMapUsers1 := `- userarn: "123456"
  username: boba.fett
  groups:
  - mercenary-admins
  - mercenary-users
- userarn: "7890"
  username: luke.skywalker
  groups:
  - jedi-admins
`
			expectedMapUsers2 := `- userarn: "123456"
  username: boba.fett
  groups:
  - mercenary-users
  - mercenary-admins
- userarn: "7890"
  username: luke.skywalker
  groups:
  - jedi-admins
`

			_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG]:\n---\n%v\n---", expectedMapUsers1)
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, awsAuthCmLookupKey, synchedAwsAuthCm)
				if err != nil {
					return "", err
				}
				//_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG]:\n---\n%v\n---", synchedAwsAuthCm.Data["mapUsers"])
				return synchedAwsAuthCm.Data["mapUsers"], nil
			}, timeout, interval).Should(Or(Equal(expectedMapUsers1), Equal(expectedMapUsers2)))
		})
	})
})
