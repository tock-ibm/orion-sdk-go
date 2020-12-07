package commands

import (
	"github.com/golang/protobuf/jsonpb"
	"io/ioutil"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.ibm.com/blockchaindb/sdk/pkg/bcdb"
	"github.ibm.com/blockchaindb/server/pkg/logger"
	"github.ibm.com/blockchaindb/server/pkg/types"
)

func waitForTxCommit(session bcdb.DBSession, txID string) (*types.TxReceipt, error) {
	p, err := session.Provenance()
	if err != nil {
		return nil, errors.Wrap(err, "error accessing provenance data")
	}
	for {
		select {
		case <-time.After(5 * time.Second):
			return nil, errors.Errorf("timeout while waiting for transaction %s to commit to BCDB", txID)

		case <-time.After(50 * time.Millisecond):
			receipt, err := p.GetTransactionReceipt(txID)
			if err == nil {
				validationInfo := receipt.GetHeader().GetValidationInfo()[receipt.GetTxIndex()]
				if validationInfo.GetFlag() == types.Flag_VALID {
					return receipt, nil
				}
				return nil, errors.Errorf("transaction [%s] is invalid, reason %s ", txID, validationInfo.GetReasonIfInvalid())
			}
		}
	}
}

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

	lg.Infof("Saved tx envelope: %s", marshalToStringOrPanic(txEnv))
	lg.Infof("Saved tx receipt, file: %s", rctFile)
	lg.Infof("Saved tx receipt, file: %s", marshalToStringOrPanic(txReceipt))

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
	lg.Infof("Loaded tx envelope: %s", env)
	lg.Infof("Loaded tx receipt, file: %s", rctFile)
	lg.Infof("Loaded tx receipt, file: %s", rct)

	return env, rct, nil
}