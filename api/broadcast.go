package api

import (
	"context"
	"io/ioutil"
	"net/http"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	"github.com/irisnet/irishub/app"
)

var txconfig = app.MakeEncodingConfig().TxConfig

func (s *Server) Broadcast(w http.ResponseWriter, r *http.Request) {
	var stdTx legacytx.StdTx
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	err = cdc.UnmarshalJSON(body, &stdTx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	txBytes, err := tx.ConvertAndEncodeStdTx(txconfig, stdTx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	client, err := rpchttp.New(s.Node, "/websocket")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}
	res, err := client.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(cdc.MustMarshalJSON(sdk.NewResponseFormatBroadcastTx(res)))
	return
}
