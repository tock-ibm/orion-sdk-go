// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"io/ioutil"
	"path"

	"github.com/hyperledger-labs/orion-server/pkg/logger"
	"github.com/hyperledger-labs/orion-server/pkg/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func marshalOrPanic(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err.Error())
	}
	return b
}

func marshalToStringOrPanic(msg proto.Message) string {
	m := jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "  ",
	}
	envStr, _ := m.MarshalToString(msg)
	return envStr
}

func saveTxEvidence(demoDir, txID string, txEnv proto.Message, txReceipt *types.TxReceipt, lg *logger.SugarLogger) error {
	envFile := path.Join(demoDir, "txs", txID+".envelope")
	err := ioutil.WriteFile(envFile, marshalOrPanic(txEnv), 0644)
	if err != nil {
		return err
	}

	rctFile := path.Join(demoDir, "txs", txID+".receipt")
	err = ioutil.WriteFile(rctFile, marshalOrPanic(txReceipt), 0644)
	if err != nil {
		return err
	}

	lg.Infof("Saved tx envelope, file: %s", envFile)

	lg.Infof("Saved tx envelope: \n%s", marshalToStringOrPanic(txEnv))
	lg.Infof("Saved tx receipt, file: %s", rctFile)
	lg.Infof("Saved tx receipt: \n%s", marshalToStringOrPanic(txReceipt))

	return nil
}

func loadTxEvidence(demoDir, txID string, lg *logger.SugarLogger) (*types.DataTxEnvelope, *types.TxReceipt, error) {
	envFile := path.Join(demoDir, "txs", txID+".envelope")
	envBytes, err := ioutil.ReadFile(envFile)
	if err != nil {
		return nil, nil, err
	}
	env := &types.DataTxEnvelope{}
	err = proto.Unmarshal(envBytes, env)
	if err != nil {
		return nil, nil, err
	}

	rctFile := path.Join(demoDir, "txs", txID+".receipt")
	rctBytes, err := ioutil.ReadFile(rctFile)
	if err != nil {
		return nil, nil, err
	}
	rct := &types.TxReceipt{}
	err = proto.Unmarshal(rctBytes, rct)
	if err != nil {
		return nil, nil, err
	}

	lg.Infof("Loaded tx envelope, file: %s", envFile)
	lg.Infof("Loaded tx envelope: \n%s", marshalToStringOrPanic(env))
	lg.Infof("Loaded tx receipt, file: %s", rctFile)
	lg.Infof("Loaded tx receipt: \n%s", marshalToStringOrPanic(rct))

	return env, rct, nil
}

func usersMap(users ...string) map[string]bool {
	m := make(map[string]bool)
	for _, u := range users {
		m[u] = true
	}
	return m
}
