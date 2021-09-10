package model_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gp42/aws-auth-operator/controllers/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type mockIamClientGetGroup func(ctx context.Context, params *iam.GetGroupInput, optFns ...func(*iam.Options)) (*iam.GetGroupOutput, error)

func (m mockIamClientGetGroup) GetGroup(ctx context.Context, params *iam.GetGroupInput, optFns ...func(*iam.Options)) (*iam.GetGroupOutput, error) {
	return m(ctx, params, optFns...)
}

var _ = Describe("MapUsers", func() {
	Describe("manipulating MapUsers object", func() {
		Context("AddUser", func() {
			It("generates correct object", func() {
				m := model.NewMapUsers()
				m.AddUser("boba.fett", "123456", "mercenary-users")
				m.AddUser("boba.fett", "123456", "mercenary-admins")

				correctGroups := map[string]struct{}{
					"mercenary-admins": {},
					"mercenary-users":  {},
				}

				Expect(m.Users["boba.fett"].UserArn).To(Equal("123456"))
				Expect(m.Users["boba.fett"].Groups).To(Equal(correctGroups))
			})
		})

		Context("LoadFromIamGroup", func() {
			It("loads valid object based on IAM group", func() {
				mock := mockIamClientGetGroup(func(ctx context.Context, params *iam.GetGroupInput, optFns ...func(*iam.Options)) (*iam.GetGroupOutput, error) {
					var r *iam.GetGroupOutput
					if params.Marker == nil {
						r = &iam.GetGroupOutput{
							IsTruncated: true,
							Marker:      aws.String("marker"),
							Group: &iamtypes.Group{
								Arn:       aws.String("123456"),
								GroupName: params.GroupName,
							},
							Users: []iamtypes.User{{
								Arn:      aws.String("123456"),
								UserName: aws.String("boba.fett"),
							}},
						}
					} else {
						r = &iam.GetGroupOutput{
							IsTruncated: false,
							Marker:      nil,
							Group: &iamtypes.Group{
								GroupName: params.GroupName,
							},
							Users: []iamtypes.User{{
								Arn:      aws.String("7890"),
								UserName: aws.String("luke.skywalker"),
							}},
						}
					}

					return r, nil
				})

				m := model.NewMapUsers()
				err := m.LoadFromIamGroup(context.TODO(), mock, "mercenary-admins", "mercenary-admins")
				Expect(err).NotTo(HaveOccurred())
				//_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG]:\n---\n%v\n---", m)

				correctGroups := map[string]struct{}{
					"mercenary-admins": {},
				}

				Expect(m.Users["boba.fett"].UserArn).To(Equal("123456"))
				Expect(m.Users["boba.fett"].Groups).To(Equal(correctGroups))
				Expect(m.Users["luke.skywalker"].UserArn).To(Equal("7890"))
				Expect(m.Users["luke.skywalker"].Groups).To(Equal(correctGroups))
			})
		})
	})

	Describe("dumping to AwsAuth", func() {
		var (
			mapUsers model.MapUsers
		)

		var _ = BeforeEach(func() {
			//logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
			mapUsers = model.NewMapUsers()
			mapUsers.AddUser("boba.fett", "123456", "mercenary-admins")
			mapUsers.AddUser("boba.fett", "123456", "mercenary-users")
		})

		Context("dumping to AwsAuthMapUsers", func() {
			It("creates a valid object", func() {
				dump, _, err := mapUsers.ToAwsAuthMapUsersDump()
				Expect(err).NotTo(HaveOccurred())
				validDump1 := `- userarn: "123456"
  username: boba.fett
  groups:
  - mercenary-users
  - mercenary-admins
`
				validDump2 := `- userarn: "123456"
  username: boba.fett
  groups:
  - mercenary-admins
  - mercenary-users
`
				//_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG] dump:\n---\n%v\n---", dump)
				Expect(dump).To(Or(Equal(validDump1), Equal(validDump2)))
			})
		})

		Context("generating patch for aws-auth", func() {
			It("creates a valid patch", func() {
				patch, _, err := mapUsers.ToPatch()
				Expect(err).NotTo(HaveOccurred())
				validPatch1 := []byte(`{"data":{"mapUsers":"- userarn: \"123456\"\n  username: boba.fett\n  groups:\n  - mercenary-admins\n  - mercenary-users\n"}}`)
				validPatch2 := []byte(`{"data":{"mapUsers":"- userarn: \"123456\"\n  username: boba.fett\n  groups:\n  - mercenary-users\n  - mercenary-admins\n"}}`)
				//_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG]:\n---\n%v\n---", patch)
				Expect(patch).To(Or(Equal(validPatch1), Equal(validPatch2)))
			})
		})

		Context("converts to awsAuthMapUsersItem", func() {
			It("creates a valid object", func() {
				user := mapUsers.Users["boba.fett"]
				item := (&user).ToAwsAuthMapUsersItem()
				validItem1 := model.AwsAuthMapUsersItem{
					UserArn:  "123456",
					UserName: "boba.fett",
					Groups:   []string{"mercenary-users", "mercenary-admins"},
				}
				validItem2 := model.AwsAuthMapUsersItem{
					UserArn:  "123456",
					UserName: "boba.fett",
					Groups:   []string{"mercenary-admins", "mercenary-users"},
				}
				Expect(item).To(Or(Equal(validItem1), Equal(validItem2)))
			})
		})
	})

})
