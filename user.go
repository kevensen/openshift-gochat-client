package main

/* {"kind":"User",
	"apiVersion":"v1",
	"metadata": {
		"name":"developer",
		"selfLink":"/oapi/v1/users/developer",
		"uid":"781595f1-fb80-11e7-a90a-e2bc220fd7c4",
		"resourceVersion":"700",
		"creationTimestamp":"2018-01-17T12:18:04Z"
	},
	"identities":["anypassword:developer"],
	"groups":[]
}*/

type userMetadata struct {
	Name              string
	SelfLink          string
	Uid               string
	ResourceVersion   string
	CreationTimestamp string
}

type User struct {
	Kind       string
	ApiVersion string
	Metadata   userMetadata
	Identities []string
	Groups     []string
}
