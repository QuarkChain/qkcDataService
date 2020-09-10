package controllers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/QuarkChain/goquarkchain/account"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"sync"
)

var (
	SDK          = new(QKCSDK)
	gasLimit     = uint64(6000000)
	gasPrice     = new(big.Int).SetUint64(1000000000)
	MaxPostLen   = (int(gasLimit) - 21000) / 68
	emptyAddress = common.Address{}
	token        = TokenIDEncode("QKC")
	fullShardID  = 262145
)

type QKCSDK struct {
	jrpcHost    jsonrpc.RPCClient
	signAccount account.Account
	mu          sync.Mutex

	nonce     uint64
	networkID uint32
}

func NewQKCSDK(privateKey, host string) *QKCSDK {
	acc, err := account.NewAccountWithKey(account.BytesToIdentityKey(common.FromHex(privateKey)))
	if err != nil {
		panic(err)
	}
	acc.QKCAddress = acc.QKCAddress.AddressInBranch(account.Branch{Value: uint32(fullShardID)})

	q := &QKCSDK{
		jrpcHost:    jsonrpc.NewClient(host),
		signAccount: acc,
	}
	q.networkID = q.GetNetWorkID()
	q.resetNonce()
	q.BalanceCheck()
	return q
}

func (q *QKCSDK) BalanceCheck() {
	fmt.Println("签名账户为", q.signAccount.QKCAddress.ToHex())
	fmt.Println("nonce:", q.nonce)

	data, err := q.GetAccountData(true)
	if err != nil {
		panic(err)
	}
	qkcValue := new(big.Int)

	bal, ok := data["balances"]
	if !ok || len(bal.([]interface{})) == 0 {

	} else {
		bList := bal.([]interface{})
		for index := 0; index < len(bList); index++ {
			b := bList[index].(map[string]interface{})

			qkcValue, err = hexutil.DecodeBig(b["balance"].(string))
			if err != nil {
				panic(err)
			}
			fmt.Println("token:", b["tokenStr"], ", balance(wei):", qkcValue)
		}
	}
	if qkcValue.Cmp(new(big.Int)) == 0 {
		panic(fmt.Errorf("account %s have no qkc", q.signAccount.QKCAddress.ToHex()))
	}
}

func (q *QKCSDK) resetNonce() {
	var err error
	tryCnt := 0
	for tryCnt <= 10 {
		tryCnt++
		q.nonce, err = q.GetNonceFromJRPC()
		if err == nil {
			return
		}
	}

}

func (q *QKCSDK) SendFormData(nonce uint64, payLoad []byte) (string, error) {
	tx := newEvmTransaction(nonce, &emptyAddress, new(big.Int), gasLimit, gasPrice, uint32(fullShardID), uint32(fullShardID), token, token, q.networkID, 0, payLoad)
	prvKey, err := crypto.ToECDSA(common.FromHex(q.signAccount.PrivateKey()))
	if err != nil {
		return "", err
	}

	tx, err = SignTx(tx, prvKey)
	if err != nil {
		return "", err
	}
	hash, err := q.SendTransaction(tx)
	if err != nil {
		return "", err
	}
	return hash, err
}

func (q *QKCSDK) SendTransaction(tx *EvmTransaction) (string, error) {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}

	resp, err := q.jrpcHost.Call("sendRawTransaction", common.ToHex(data))
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	return resp.Result.(string), nil
}

func (q *QKCSDK) GetNonceFromJRPC() (nonce uint64, err error) {
	shrd, err := q.GetAccountData(true)
	if err != nil {
		return 0, err
	}
	return hexutil.DecodeUint64(shrd["transactionCount"].(string))
}

func (q *QKCSDK) GetNonce() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.nonce++
	return q.nonce - 1
}
func GetFullShardIdByFullShardKey(fullShardKey uint32) uint32 {
	chainID := fullShardKey >> 16
	shardsize := uint32(1)
	shardID := fullShardKey & (shardsize - 1)
	return (chainID << 16) | shardsize | shardID
}

func (q *QKCSDK) GetAccountData(includeShards bool) (map[string]interface{}, error) {
	resp, err := q.jrpcHost.Call("getAccountData", q.signAccount.QKCAddress.ToHex(), nil, includeShards)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	fullShardId := GetFullShardIdByFullShardKey(q.signAccount.QKCAddress.FullShardKey)
	shards := resp.Result.(map[string]interface{})["shards"]
	for _, val := range shards.([]interface{}) {
		shrd := val.(map[string]interface{})
		id, err := hexutil.DecodeUint64(shrd["fullShardId"].(string))
		if err != nil {
			return nil, err
		}
		if id == uint64(fullShardId) {
			return shrd, nil
		}
	}
	return nil, errors.New("has no such account")
}

func (q *QKCSDK) GetTransactionById(txid string) (result []byte, err error) {
	resp, err := q.jrpcHost.Call("getTransactionById", []string{txid})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	if _, ok := resp.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("txid %v not in the chain", txid)
	}
	return hex.DecodeString(resp.Result.(map[string]interface{})["data"].(string)[2:])
}

func (q *QKCSDK) GetNetWorkID() uint32 {
	resp, err := q.jrpcHost.Call("networkInfo")
	if err != nil {
		panic(err)
	}
	if resp.Error != nil {
		panic(resp.Error)
	}
	networkId, err := hexutil.DecodeUint64(resp.Result.(map[string]interface{})["networkId"].(string))
	if err != nil {
		panic(err)
	}
	return uint32(networkId)
}
