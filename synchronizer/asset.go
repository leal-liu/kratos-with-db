package synchronizer

import (
	"encoding/json"
	"fmt"
	"time"
)

// AssetSync
type AssetSync struct {
	*Base
}

// Sync sync user coins
func (object *AssetSync) Sync(user, denom string,
	timeout time.Duration) (coins []string, err error) {
	object.SetUri(fmt.Sprintf("http://%s:%d/assets/coins/%s",
		object.host,
		object.port,
		user))
	var code int
	var raw []byte
	if code, raw, err = object.Base.Sync(timeout); nil != err {
		return
	}
	if 200 > code || 299 < code {
		return
	}
	type Response struct {
		Height string `json:"height"`
		Result []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"result"`
	}
	var response Response
	if err = json.Unmarshal(raw, &response); nil != err {
		return
	}
	for _, result := range response.Result {
		if denom != result.Denom {
			continue
		}
		coins = append(coins, result.Amount)
	}
	return
}

func NewAssetSync(host string, port int) *AssetSync {
	object := &AssetSync{
		Base: &Base{host: host, port: port},
	}
	return object
}
