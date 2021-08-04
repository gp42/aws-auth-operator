/*
MapUsers represents a 'mapUsers' section of aws-auth configmap:
https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html
*/

package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"gopkg.in/yaml.v2"

	"github.com/gp42/aws-auth-operator/controllers/common"
)

type MapUsers struct {
	Users map[string]MapUsersItem
}

func NewMapUsers() MapUsers {
	return MapUsers{
		Users: map[string]MapUsersItem{},
	}
}

func (m *MapUsers) AddUser(name, arn, group string) {
	userGroups := map[string]struct{}{}
	if val, ok := m.Users[name]; ok {
		userGroups = val.Groups
		userGroups[group] = struct{}{}
	} else {
		userGroups[group] = struct{}{}
	}
	m.Users[name] = MapUsersItem{
		UserArn:  arn,
		UserName: name,
		Groups:   userGroups,
	}
}

func (m *MapUsers) ToAwsAuthMapUsersDump() (string, string, error) {
	awsAuthMapUsers := AwsAuthMapUsers{}
	for _, v := range m.Users {
		awsAuthMapUsers.MapUsers = append(awsAuthMapUsers.MapUsers, v.ToAwsAuthMapUsersItem())
	}

	awsAuthMapUsersDump, err := yaml.Marshal(&awsAuthMapUsers.MapUsers)
	if err != nil {
		return "", "", err
	}

	stringSum := common.Sha512SumFromStringSorted(string(awsAuthMapUsersDump))

	return string(awsAuthMapUsersDump), stringSum, nil
}

func (m *MapUsers) ToPatch() ([]byte, string, error) {
	mapUsersDump, stringSum, err := m.ToAwsAuthMapUsersDump()
	if err != nil {
		return nil, "", err
	}

	patch := AwsAuthPatch{
		Data: AwsAuthData{
			MapUsers: string(mapUsersDump),
		},
	}

	patchDump, err := json.Marshal(&patch)
	if err != nil {
		return nil, "", err
	}

	return patchDump, stringSum, nil
}

type IamClientInterface interface {
	GetGroup(context.Context, *iam.GetGroupInput, ...func(*iam.Options)) (*iam.GetGroupOutput, error)
}

func (m *MapUsers) LoadFromIamGroup(ctx context.Context, iamClient IamClientInterface, groupSource, groupDest string) error {
	var marker *string = nil
	for {
		iamGroupOutput, err := iamClient.GetGroup(ctx,
			&iam.GetGroupInput{
				GroupName: aws.String(groupSource),
				Marker:    marker,
				MaxItems:  aws.Int32(1),
			})
		if err != nil {
			return fmt.Errorf("Unable to get iam group: %w", err)
		}
		for _, iamUser := range iamGroupOutput.Users {
			m.AddUser(*iamUser.UserName, *iamUser.Arn, groupDest)
		}
		if iamGroupOutput.IsTruncated {
			marker = iamGroupOutput.Marker
			continue
		}
		break
	}

	return nil
}

type MapUsersItem struct {
	UserArn  string              `json:"userarn"`
	UserName string              `json:"username"`
	Groups   map[string]struct{} `json:"groups"`
}

func (m *MapUsersItem) ToAwsAuthMapUsersItem() AwsAuthMapUsersItem {
	groups := []string{}
	for k, _ := range m.Groups {
		groups = append(groups, k)
	}
	return AwsAuthMapUsersItem{
		UserArn:  m.UserArn,
		UserName: m.UserName,
		Groups:   groups,
	}
}
