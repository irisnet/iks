package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	ckeys "github.com/irisnet/irishub/client/keys"
	"github.com/stretchr/testify/require"
)

const (
	sMenominc = "marine intact tone element chest certain school village sound guilt nothing deposit cart skirt unveil bulk unit dust peasant cannon faith lyrics swear regret"
	sAcc      = "faa1yv6alpum5r0nmnzkk4esp3cs5d58h8g9ek2ekr"
	sAccPub   = "fap1addwnpepqwf7vcxxzylpk4lgghqeyqdv4dwhh0htk5ukskqr7p383t8udzdkgwgjl09"
	sVal      = "fva1yv6alpum5r0nmnzkk4esp3cs5d58h8g9v8qkty"
	sValPub   = "fvp1addwnpepqwf7vcxxzylpk4lgghqeyqdv4dwhh0htk5ukskqr7p383t8udzdkgddnsmw"
	sCons     = "fca1yv6alpum5r0nmnzkk4esp3cs5d58h8g99lh7f8"
	sConsPub  = "fcp1addwnpepqwf7vcxxzylpk4lgghqeyqdv4dwhh0htk5ukskqr7p383t8udzdkgljacuh"

	testKey     = "jack"
	testPass    = "123456789"
	testPassAlt = "foobarbaz"
)

func TestGetKeys(t *testing.T) {
	server := setup(t)
	defer server.Close()

	// test empty get keys
	expected := fmt.Sprintf("[]")
	empty := getRoute(t, fmt.Sprintf("%s/keys", server.URL), 200)
	require.Equal(t, expected, string(empty))

	// test adding a key w/ existing mnemonic
	addNP := AddNewKey{Name: testKey, Password: testPass, Mnemonic: sMenominc}
	key := unmarshalKeyOutput(postRoute(t, fmt.Sprintf("%s/keys", server.URL), addNP.Marshal(), 200))
	require.Equal(t, addNP.Name, key.Name)
	require.Equal(t, sAcc, key.Address)
	require.Equal(t, sAccPub, key.PubKey)

	// test invalid key type
	getRoute(t, fmt.Sprintf("%s/keys/foo?bech=foo", server.URL), 400)

	// test key not exists
	getRoute(t, fmt.Sprintf("%s/keys/foo?bech=acc", server.URL), 404)

	// test return bech val prefix
	valKey := unmarshalKeyOutput(getRoute(t, fmt.Sprintf("%s/keys/jack?bech=val", server.URL), 200))
	require.Equal(t, sVal, valKey.Address)
	require.Equal(t, sValPub, valKey.PubKey)

	// test return bech cons prefix
	consKey := unmarshalKeyOutput(getRoute(t, fmt.Sprintf("%s/keys/jack?bech=cons", server.URL), 200))
	require.Equal(t, sCons, consKey.Address)
	require.Equal(t, sConsPub, consKey.PubKey)

	// TestGetKeys
	keys := unmarshalKeysOutput(getRoute(t, fmt.Sprintf("%s/keys", server.URL), 200))
	require.Equal(t, len(keys), 1)
	require.Equal(t, testKey, keys[0].Name)

	// TestUpdateKey bad path
	badUpdatePass := UpdateKeyBody{OldPassword: testKey, NewPassword: testPassAlt}
	wrongPass := unmarshalError(putRoute(t, fmt.Sprintf("%s/keys/%s", server.URL, testKey), badUpdatePass.Marshal(), 401))
	require.NotEmpty(t, wrongPass.Error)

	// TestUpdateKey happy path
	updatePass := UpdateKeyBody{OldPassword: testPass, NewPassword: testPassAlt}
	goodPass := putRoute(t, fmt.Sprintf("%s/keys/%s", server.URL, testKey), updatePass.Marshal(), 204)
	require.Empty(t, goodPass)

	// Test delete key bad path
	deleteKey := DeleteKeyBody{Password: testPass}
	badPath := unmarshalError(deleteRoute(t, fmt.Sprintf("%s/keys/%s", server.URL, testKey), deleteKey.Marshal(), 401))
	require.NotEmpty(t, badPath.Error)

	// Test delete key happy path
	deleteKey = DeleteKeyBody{Password: testPassAlt}
	happyPath := deleteRoute(t, fmt.Sprintf("%s/keys/%s", server.URL, testKey), deleteKey.Marshal(), 200)
	require.Empty(t, happyPath)
}

func unmarshalError(in []byte) (out restError) {
	err := json.Unmarshal(in, &out)
	if err != nil {
		panic(err)
	}
	return
}

func unmarshalKeyOutput(ko []byte) (out ckeys.KeyOutput) {
	err := json.Unmarshal(ko, &out)
	if err != nil {
		panic(err)
	}
	return
}

func unmarshalKeysOutput(ko []byte) (out []ckeys.KeyOutput) {
	err := json.Unmarshal(ko, &out)
	if err != nil {
		panic(err)
	}
	return
}

func setup(t *testing.T) *httptest.Server {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{KeyDir: dir}
	return httptest.NewServer(s.Router())
}

func getRoute(t *testing.T, route string, expStatus int) []byte {
	resp, err := http.Get(route)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != expStatus {
		t.Fatalf("Expected status '%d', got '%d'\n", expStatus, resp.StatusCode)
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func postRoute(t *testing.T, route string, data []byte, expStatus int) []byte {
	resp, err := http.Post(route, "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != expStatus {
		t.Fatalf("Expected status '%d', got '%d'\n", expStatus, resp.StatusCode)
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func putRoute(t *testing.T, route string, data []byte, expStatus int) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, route, bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != expStatus {
		t.Fatalf("Expected status '%d', got '%d'\n", expStatus, resp.StatusCode)
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func deleteRoute(t *testing.T, route string, data []byte, expStatus int) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, route, bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != expStatus {
		t.Fatalf("Expected status '%d', got '%d'\n", expStatus, resp.StatusCode)
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
