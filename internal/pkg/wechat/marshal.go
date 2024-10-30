package wechat

import (
	"encoding/json"
	"fmt"
	"github.com/xen0n/go-workwx"
)

func Marshal(oaDetail workwx.OAApprovalDetail) (map[string]interface{}, error) {
	evtData, err := json.Marshal(oaDetail)
	if err != nil {
		fmt.Println("Error marshalling struct:", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(evtData, &data)
	return data, err
}

func Unmarshal(data map[string]interface{}) (workwx.OAApprovalDetail, error) {
	evtData, err := json.Marshal(data)
	if err != nil {
		return workwx.OAApprovalDetail{}, err
	}

	var oaDetail workwx.OAApprovalDetail
	err = json.Unmarshal(evtData, &oaDetail)
	return oaDetail, err
}
