/*
AwsAuthPatch represents a patch to aws-auth configmap:
https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html
*/

package model

type AwsAuthPatch struct {
	Data AwsAuthData `json:"data"`
}

type AwsAuthData struct {
	MapUsers string `json:"mapUsers"`
}

type AwsAuthMapUsers struct {
	MapUsers []AwsAuthMapUsersItem `json:"mapUsers"`
}

type AwsAuthMapUsersItem struct {
	UserArn  string   `json:"userarn"`
	UserName string   `json:"username"`
	Groups   []string `json:"groups"`
}
