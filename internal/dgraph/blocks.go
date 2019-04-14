package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// GetBlockHashFromHeight returnes the hash of the block retrieving it based on its height
func GetBlockHashFromHeight(height int32) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: eq(height, %d), first: 1) {
			block
		}
	}`, height))
	if err != nil {
		return "", err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.Q) == 0 {
		return "", errors.New("No address occurences")
	}
	return r.Q[0].Block, nil
}
