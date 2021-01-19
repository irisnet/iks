package api

import (
	"context"
	"errors"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/irisnet/irishub/app"
)

const (
	maxValidAccountValue = int(0x80000000 - 1)
	maxValidIndexalue    = int(0x80000000 - 1)
)

var cdc *codec.LegacyAmino

func init() {
	_, cdc = app.MakeCodecs()
	cdc.RegisterInterface((*Info)(nil), nil)
	cdc.RegisterConcrete(localInfo{}, "crypto/keys/localInfo", nil)
	cdc.RegisterConcrete(ledgerInfo{}, "crypto/keys/ledgerInfo", nil)
	cdc.RegisterConcrete(offlineInfo{}, "crypto/keys/offlineInfo", nil)
	cdc.RegisterConcrete(multiInfo{}, "crypto/keys/multiInfo", nil)
}

// Server represents the API server
type Server struct {
	Port   int    `json:"port"`
	KeyDir string `json:"key_dir"`
	Node   string `json:"node"`

	Version string `yaml:"version,omitempty"`
	Commit  string `yaml:"commit,omitempty"`
	Branch  string `yaml:"branch,omitempty"`
}

// Router returns the router
func (s *Server) Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/version", s.VersionHandler).Methods("GET")
	router.HandleFunc("/keys", s.GetKeys).Methods("GET")
	router.HandleFunc("/keys", s.PostKeys).Methods("POST")
	router.HandleFunc("/keys/{name}", s.GetKey).Methods("GET")
	router.HandleFunc("/keys/{name}", s.PutKey).Methods("PUT")
	router.HandleFunc("/keys/{name}", s.DeleteKey).Methods("DELETE")
	router.HandleFunc("/tx/sign", s.Sign).Methods("POST")
	router.HandleFunc("/tx/broadcast", s.Broadcast).Methods("POST")
	router.HandleFunc("/tx/bank/send", s.BankSend).Methods("POST")

	return router
}

// SimulateGas simulates gas for a transaction
func (s *Server) SimulateGas(txbytes []byte) (res uint64, err error) {
	client, err := rpchttp.New(s.Node, "/websocket")
	if err != nil {
		return
	}
	result, err := client.ABCIQueryWithOptions(context.Background(),
		"/app/simulate",
		tmbytes.HexBytes(txbytes),
		rpcclient.ABCIQueryOptions{},
	)

	if err != nil {
		return
	}

	if !result.Response.IsOK() {
		return 0, errors.New(result.Response.Log)
	}

	var simulationResult sdk.SimulationResponse
	if err := jsonpb.Unmarshal(strings.NewReader(string(result.Response.Value)), &simulationResult); err != nil {
		return 0, err
	}

	return simulationResult.GasUsed, nil
}
