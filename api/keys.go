package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	bip39 "github.com/cosmos/go-bip39"
)

var (
	KeyringServiceName = "iris"
	Kb                 LegacyKeybase
)

// GetKeys is the handler for the GET /keys
func (s *Server) GetKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	infos, err := Kb.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	if len(infos) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	keysOutput, err := cryptokeyring.Bech32KeysOutput(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	out, err := json.Marshal(keysOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return
}

// AddNewKey is the necessary data for adding a new key
type AddNewKey struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Mnemonic string `json:"mnemonic,omitempty"`
	Account  int    `json:"account,string,omitempty"`
	Index    int    `json:"index,string,omitempty"`
}

func (ak AddNewKey) Marshal() []byte {
	out, err := json.Marshal(ak)
	if err != nil {
		panic(err)
	}
	return out
}

// PostKeys is the handler for the POST /keys
func (s *Server) PostKeys(w http.ResponseWriter, r *http.Request) {
	var m AddNewKey

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	err = json.Unmarshal(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	if m.Name == "" || m.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("must include both password and name with request")).marshal())
		return
	}

	// if mnemonic is empty, generate one
	mnemonic := m.Mnemonic
	if mnemonic == "" {
		_, mnemonic, _ = cryptokeyring.NewInMemory().NewMnemonic("inmemorykey", cryptokeyring.English, FullFundraiserPath, hd.Secp256k1)
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("invalid mnemonic %s", mnemonic)).marshal())
		return
	}

	if m.Account < 0 || m.Account > maxValidAccountValue {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("invalid account number")).marshal())
		return
	}

	if m.Index < 0 || m.Index > maxValidIndexalue {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("invalid index number")).marshal())
		return
	}

	_, err = Kb.Get(m.Name)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(fmt.Errorf("key %s already exists", m.Name)).marshal())
		return
	}

	// account := uint32(m.Account)
	// index := uint32(m.Index)
	info, err := Kb.CreateKey(m.Name, mnemonic, m.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	keyOutput, err := cryptokeyring.Bech32KeyOutput(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	keyOutput.Mnemonic = mnemonic

	out, err := json.Marshal(keyOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return
}

// GetKey is the handler for the GET /keys/{name}
func (s *Server) GetKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["name"]
	bechPrefix := r.URL.Query().Get("bech")

	if bechPrefix == "" {
		bechPrefix = "acc"
	}

	bechKeyOut, err := getBechKeyOut(bechPrefix)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	info, err := Kb.Get(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	keyOutput, err := bechKeyOut(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	out, err := json.Marshal(keyOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return
}

type bechKeyOutFn func(keyInfo cryptokeyring.Info) (cryptokeyring.KeyOutput, error)

func getBechKeyOut(bechPrefix string) (bechKeyOutFn, error) {
	switch bechPrefix {
	case "acc":
		return cryptokeyring.Bech32KeyOutput, nil
	case "val":
		return cryptokeyring.Bech32ValKeyOutput, nil
	case "cons":
		return cryptokeyring.Bech32ConsKeyOutput, nil
	}

	return nil, fmt.Errorf("invalid Bech32 prefix encoding provided: %s", bechPrefix)
}

// UpdateKeyBody update key password request REST body
type UpdateKeyBody struct {
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
}

func (u UpdateKeyBody) Marshal() []byte {
	out, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}
	return out
}

// PutKey is the handler for the PUT /keys/{name}
func (s *Server) PutKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var m UpdateKeyBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	err = Kb.Update(name, m.OldPassword, func() (string, error) { return m.NewPassword, nil })
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

// DeleteKeyBody request
type DeleteKeyBody struct {
	Password string `json:"password"`
}

func (u DeleteKeyBody) Marshal() []byte {
	out, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}
	return out
}

// DeleteKey is the handler for the DELETE /keys/{name}
func (s *Server) DeleteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var m DeleteKeyBody

	kb, err := NewLegacy(KeyringServiceName, s.KeyDir, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(newError(err).marshal())
		return
	}

	err = kb.Delete(name, m.Password, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(newError(err).marshal())
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
