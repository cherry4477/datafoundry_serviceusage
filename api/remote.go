package api

import (
	"os"
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/asiainfoLDP/datahub_commons/common"
)

//================================================================
// 
//================================================================

var (
	DF_HOST     string
	DF_API_Auth string
)

func init() {
	DF_HOST = os.Getenv("DATAFOUNDRY_HOST_ADDR")
	DF_API_Auth = DF_HOST + "/oapi/v1/users/~"

	Logger.Info("DF_HOST = ", DF_HOST)
	Logger.Info("DF_API_Auth = ", DF_API_Auth)
}

//================================================================
// 
//================================================================

type ObjectMeta struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

type User struct {
	ObjectMeta `json:"metadata,omitempty"`

	// FullName is the full name of user
	FullName string `json:"fullName,omitempty"`

	// Identities are the identities associated with this user
	Identities []string `json:"identities"`

	// Groups are the groups that this user is a member of
	Groups []string `json:"groups"`
}

func authDF(token string) (*User, error) {
	response, data, err := common.RemoteCall("GET", DF_API_Auth, token, "")
	if err != nil {
		Logger.Debugf("authDF error: ", err.Error())
		return nil, err
	}

	// todo: use return code and msg instead
	if response.StatusCode != http.StatusOK {
		Logger.Debugf("remote (%s) status code: %d. data=%s", DF_API_Auth, response.StatusCode, string(data))
		return nil, fmt.Errorf("remote (%s) status code: %d.", DF_API_Auth, response.StatusCode)
	}

	user := new(User)
	err = json.Unmarshal(data, user)
	if err != nil {
		Logger.Debugf("authDF Unmarshal error: %s. Data: %s\n", err.Error(), string(data))
		return nil, err
	}

	return user, nil
}

func dfUser(user *User) string {
	return user.Name
}

func getDFUserame(token string) (string, error) {
	//Logger.Info("token = ", token)

	user, err := authDF(token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}

//=======================================================================
// 
//=======================================================================

// todo: check user permission on project

//=======================================================================
// 
//=======================================================================

// todo: send consume money request
