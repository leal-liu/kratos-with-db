package synchronizer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KuChainNetwork/kuchain/chain/types"
)

type DistributionSync struct {
	*Base
}

func (object *DistributionSync) Sync(validator, delegator string,
	timeout time.Duration) (coins types.DecCoins, err error) {
	object.SetUri(fmt.Sprintf("http://%s:%d/distribution/delegators/%s/rewards/%s",
		object.host,
		object.port,
		delegator,
		validator))
	var code int
	var raw []byte
	if code, raw, err = object.Base.Sync(timeout); nil != err {
		return
	}
	if 200 > code || 299 < code {
		return
	}
	type Response struct {
		Height string         `json:"height"`
		Result types.DecCoins `json:"result"`
	}
	var response Response
	if err = json.Unmarshal(raw, &response); nil != err {
		return
	}
	coins = response.Result
	return
}

func NewDistributionSync(host string, port int) *DistributionSync {
	return &DistributionSync{
		Base: &Base{host: host, port: port},
	}
}
