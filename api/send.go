package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/irisnet/irishub/client"
	cbank "github.com/irisnet/irishub/client/bank"
	"github.com/irisnet/irishub/modules/auth"
	sdk "github.com/irisnet/irishub/types"
)

// BankSendBody contains the necessary data to make a send transaction
type BankSendBody struct {
	Sender        sdk.AccAddress `json:"sender"`
	Reciever      sdk.AccAddress `json:"reciever"`
	Amount        string         `json:"amount"`
	ChainID       string         `json:"chain-id"`
	Memo          string         `json:"memo,omitempty"`
	Fees          string         `json:"fees,omitempty"`
	GasAdjustment string         `json:"gas_adjustment,omitempty"`
}

func (sb BankSendBody) Marshal() []byte {
	out, err := json.Marshal(sb)
	if err != nil {
		panic(err)
	}
	return out
}

// BankSend handles the /tx/bank/send route
func (s *Server) BankSend(w http.ResponseWriter, r *http.Request) {
	var sb BankSendBody

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	err = cdc.UnmarshalJSON(body, &sb)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	coins, err := sdk.ParseCoins(sb.Amount)

	coins, err = Convert2MinUnitCoins(coins)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("failed to parse amount %s into sdk.Coins", sb.Amount)).marshal())
		return
	}
	sb.Amount = coins.String()

	var fees sdk.Coin
	if sb.Fees != "" {
		fees, err = sdk.ParseCoin(sb.Fees)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(newError(fmt.Errorf("failed to parse fees %s into sdk.Coins", sb.Fees)).marshal())
			return
		}
		if fees.Denom == sdk.IRIS.Name {
			fees, err = sdk.IRIS.ConvertToMinCoin(fees.String())
		}
		sb.Fees = fees.String()
	}

	stdTx := auth.NewStdTx(
		[]sdk.Msg{cbank.BuildBankSendMsg(sb.Sender, sb.Reciever, coins)},
		auth.NewStdFee(20000, fees),
		[]auth.StdSignature{{}},
		sb.Memo,
	)

	//gas, err := s.SimulateGas(cdc.MustMarshalBinaryLengthPrefixed(stdTx))
	// always use default gas
	gas := uint64(client.DefaultGasLimit)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	if sb.GasAdjustment == "" {
		sb.GasAdjustment = "1"
	}
	//if gas != 0 {
	//	adj, err := strconv.ParseFloat(sb.GasAdjustment, 64)
	//	if err != nil {
	//		w.WriteHeader(http.StatusBadRequest)
	//		w.Write(newError(fmt.Errorf("failed to parse gasAdjustment %d into float64", sb.GasAdjustment)).marshal())
	//		return
	//	}
	//	gas = uint64(adj * float64(gas))
	//}

	adj, err := strconv.ParseFloat(sb.GasAdjustment, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("failed to parse gasAdjustment %d into float64", sb.GasAdjustment)).marshal())
		return
	}
	gas = uint64(adj * float64(gas))

	stdTx = auth.NewStdTx(
		stdTx.Msgs,
		auth.NewStdFee(gas, fees),
		[]auth.StdSignature{},
		stdTx.Memo,
	)

	w.WriteHeader(http.StatusOK)
	w.Write(cdc.MustMarshalJSON(stdTx))
	return
}

func Convert2MinUnitCoins(coins sdk.Coins) (mincoins sdk.Coins, err error) {
	for i, coin := range coins {
		if coin.Denom == sdk.IRIS.Name {
			coins[i], err = sdk.IRIS.ConvertToMinCoin(coin.String())
			if err != nil {
				return coins, err
			}
		}
	}
	return coins, err
}
