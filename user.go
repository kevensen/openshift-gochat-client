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
	token      string
	Kind       string
	ApiVersion string
	Metadata   userMetadata
	Identities []string
	Groups     []string
}

func (user *User) login() (error, int) {

	/*glog.Infoln("Obtaining user from", *OpenshiftApiHost+"/oapi/v1/users/~")
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(user.token).
		Get("https://" + *OpenshiftApiHost + "/oapi/v1/users/~")

	glog.Infoln("Status Code:", resp.StatusCode())

	if err != nil {
		return err, resp.StatusCode()
	}
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		return err, resp.StatusCode()
	}*/
	//return nil, resp.StatusCode()
	return nil, 200
}
