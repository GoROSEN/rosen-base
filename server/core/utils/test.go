package utils

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CheckJSONResult 检查返回的json是否成功（测试专用）
func CheckJSONResult(body []byte) (map[string]interface{}, error) {

	// Check the response body is what we expect.
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Unmarshal failed, ", err)
		return nil, err
	}
	// expected := `{"code":200,"data":{"items":[],"pager":{"page":1,"pageCount":0,"pageSize":20}},"msg":"OK"}`
	code, ok := m["code"].(float64)
	if !ok {
		return nil, errors.New("expecting code in json, got nil")
	}
	if int(code) != 200 {
		return nil, fmt.Errorf("expect code = 200, got %v", code)
	}
	return m, nil
}
