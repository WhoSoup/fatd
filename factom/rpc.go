package factom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	_log "bitbucket.org/canonical-ledgers/fatd/log"

	jrpc "github.com/AdamSLevy/jsonrpc2/v3"
)

var log _log.Log

func Init() {
	log = _log.New("factom")
}

func request(method string, params interface{}, result interface{}) error {
	id := rand.Uint32()%200 + 500
	reqBytes, err := json.Marshal(jrpc.NewRequest(method, id, params))
	if err != nil {
		return fmt.Errorf("json.Marshal(jrpc.NewRequest(%#v, %v, %#v): %v",
			method, id, params, err)
	}
	//log.Debugf("%v", string(reqBytes))
	endpoint := "http://" + RpcConfig.FactomdServer + "/v2"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("http.NewRequest(%#v, %#v, %#v): %v",
			http.MethodPost, endpoint, reqBytes, err)
	}
	req.Header.Add("Content-Type", "application/json")

	c := http.Client{Timeout: RpcConfig.FactomdTimeout}
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client%#v.Do(%#v): %v",
			c, req, err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %#v", res.Status)
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll(http.Response.Body): %v", err)
	}
	//log.Debugf("%v", string(resBytes))

	resJrpc := jrpc.NewResponse(result)
	if err := json.Unmarshal(resBytes, resJrpc); err != nil {
		return fmt.Errorf("json.Unmarshal(, ): %v", err)
	}
	if resJrpc.Error != nil {
		return fmt.Errorf("%#v", resJrpc.Error)
	}
	//log.Debugf("%v", resJrpc)
	return nil
}