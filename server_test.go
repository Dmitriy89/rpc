package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestRPC(t *testing.T) {

	tType := []struct {
		req, resp string
	}{
		{`{"jsonrpc":"2.0","id":1,"method":"Data.Add","params":[{"uuid": "a82ef2ce-90b9-4e27-9ba2-3b2daf4c14d9", "login": "User1"}]}`, `{"id":1,"result":"Data added to database: uuid - a82ef2ce-90b9-4e27-9ba2-3b2daf4c14d9 / login - User1","error":null}`},
		{`{"jsonrpc":"2.0","id":2,"method":"Data.Add","params":[{"uuid": "00ae9091-e16c-4909-8de4-fdb7c886174f", "login": "User2"}]}`, `{"id":2,"result":"Data added to database: uuid - 00ae9091-e16c-4909-8de4-fdb7c886174f / login - User2","error":null}`},
		{`{"jsonrpc":"2.0","id":3,"method":"Data.Add","params":[{"uuid": "5b91a34a-6223-48a6-a602-baf98a729046", "login": "User3"}]}`, `{"id":3,"result":"Data added to database: uuid - 5b91a34a-6223-48a6-a602-baf98a729046 / login - User3","error":null}`},
		{`{"jsonrpc":"2.0","id":1,"method":"Data.Get","params":[{"login": "User1"}]}`, `{"id":1,"result":"Report: a82ef2ce-90b9-4e27-9ba2-3b2daf4c14d9 / 2018.28.10","error":null}`},
		{`{"jsonrpc":"2.0","id":1,"method":"Data.Set","params":[{"uuid":"00ae9091-e16c-4909-8de4-fdb7c886174f","update":[{"newlogin":"User16"}]}]}`, `{"id":1,"result":"Data update","error":null}`},
	}

	for _, val := range tType {

		resp, err := http.Post("http://localhost:8080/rpc", "application/json", bytes.NewBufferString(val.req))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		b := strings.Trim(string(body), "\n")

		if b != val.resp {
			t.Error("Fail test")
		}
	}
}
