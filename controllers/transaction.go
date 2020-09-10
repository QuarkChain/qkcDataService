// Modified from go-ethereum under GNU Lesser General Public License

package controllers

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"math/big"
	"strings"
	"sync/atomic"
)

const (
	EvmTx = 0
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

var (
	ErrInvalidSig     = errors.New("invalid transaction v, r, s values")
	prefixOfRlpUint32 = byte(0x84)
	lenOfRlpUint32    = 5
)

const (
	tokenBase = int64(36)
)

func TokenIDEncode(str string) uint64 {
	if len(str) >= 13 {
		panic(errors.New("name too long"))
	}

	str = strings.ToUpper(str)
	id := tokenCharEncode(str[len(str)-1])
	base := uint64(tokenBase)

	for index := len(str) - 2; index >= 0; index-- {
		id += base * (tokenCharEncode(str[index]) + 1)
		base *= uint64(tokenBase)
	}
	return id
}
func tokenCharEncode(char byte) uint64 {
	if char >= byte('A') && char <= byte('Z') {
		return 10 + uint64(char-byte('A'))
	}
	if char >= byte('0') && char <= byte('9') {
		return uint64(char - byte('0'))
	}
	panic(fmt.Errorf("unknown character %v", char))
}

type Uint32 uint32

func (u *Uint32) getValue() uint32 {
	return uint32(*u)
}

func (u *Uint32) EncodeRLP(w io.Writer) error {
	bytes := make([]byte, lenOfRlpUint32)
	bytes[0] = prefixOfRlpUint32
	binary.BigEndian.PutUint32(bytes[1:], uint32(*u))
	_, err := w.Write(bytes)
	return err
}

func (u *Uint32) DecodeRLP(s *rlp.Stream) error {
	data, err := s.Raw()
	if err != nil {
		return err
	}
	if len(data) != lenOfRlpUint32 {
		return fmt.Errorf("len is %v should %v", len(data), lenOfRlpUint32)
	}

	if data[0] != prefixOfRlpUint32 {
		return fmt.Errorf("preString is wrong, is %v should %v", data[0], lenOfRlpUint32)

	}

	*u = Uint32(binary.BigEndian.Uint32(data[1:]))
	return nil
}

type EvmTransaction struct {
	data txdata
	// caches
	updated       bool
	hash          atomic.Value
	size          atomic.Value
	from          atomic.Value
	FromShardsize uint32
	ToShardsize   uint32
}

type txdata struct {
	AccountNonce     uint64          `json:"nonce"              gencodec:"required"`
	Price            *big.Int        `json:"gasPrice"           gencodec:"required"`
	GasLimit         uint64          `json:"gas"                gencodec:"required"`
	Recipient        *common.Address `json:"to"                 rlp:"nil"` // nil means contract creation
	Amount           *big.Int        `json:"value"              gencodec:"required"`
	Payload          []byte          `json:"data"            	gencodec:"required"`
	NetworkId        uint32          `json:"networkId"          gencodec:"required"`
	FromFullShardKey *Uint32         `json:"fromFullShardKey"   gencodec:"required"`
	ToFullShardKey   *Uint32         `json:"toFullShardKey"     gencodec:"required"`
	GasTokenID       uint64          `json:"gasTokenId"    		gencodec:"required"`
	TransferTokenID  uint64          `json:"transferTokenId"    gencodec:"required"`
	Version          uint32          `json:"version"            gencodec:"required"`
	// Signature values
	V *big.Int `json:"v"             gencodec:"required"`
	R *big.Int `json:"r"             gencodec:"required"`
	S *big.Int `json:"s"             gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"-"              rlp:"-"`
}

func (e *EvmTransaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(&e.data)
}

func newEvmTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, fromFullShardKey uint32, toFullShardKey uint32, gasTokenID uint64, transferTokenID uint64, networkId uint32, version uint32, data []byte) *EvmTransaction {
	newFromFullShardKey := Uint32(fromFullShardKey)
	newToFullShardKey := Uint32(toFullShardKey)
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce:     nonce,
		Recipient:        to,
		Payload:          data,
		Amount:           new(big.Int),
		GasLimit:         gasLimit,
		Price:            new(big.Int),
		FromFullShardKey: &newFromFullShardKey,
		ToFullShardKey:   &newToFullShardKey,
		GasTokenID:       gasTokenID,
		TransferTokenID:  transferTokenID,
		NetworkId:        networkId,
		Version:          version,
		V:                new(big.Int),
		R:                new(big.Int),
		S:                new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &EvmTransaction{data: d}
}

// EncodeRLP implements rlp.Encoder
func (tx *EvmTransaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *EvmTransaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

type txdataUnsigned struct {
	AccountNonce     uint64          `json:"nonce"              gencodec:"required"`
	Price            *big.Int        `json:"gasPrice"           gencodec:"required"`
	GasLimit         uint64          `json:"gas"                gencodec:"required"`
	Recipient        *common.Address `json:"to"                 rlp:"nil"` // nil means contract creation
	Amount           *big.Int        `json:"value"              gencodec:"required"`
	Payload          []byte          `json:"input"              gencodec:"required"`
	NetworkId        uint32          `json:"networkid"          gencodec:"required"`
	FromFullShardKey *Uint32         `json:"fromfullshardid"    gencodec:"required"`
	ToFullShardKey   *Uint32         `json:"tofullshardid"      gencodec:"required"`
	GasTokenID       uint64          `json:"gasTokenID"      gencodec:"required"`
	TransferTokenID  uint64          `json:"transferTokenID"      gencodec:"required"`
}

func (tx *EvmTransaction) getUnsignedHash() common.Hash {
	unsigntx := txdataUnsigned{
		AccountNonce:     tx.data.AccountNonce,
		Price:            tx.data.Price,
		GasLimit:         tx.data.GasLimit,
		Recipient:        tx.data.Recipient,
		Amount:           tx.data.Amount,
		Payload:          tx.data.Payload,
		NetworkId:        tx.data.NetworkId,
		FromFullShardKey: tx.data.FromFullShardKey,
		ToFullShardKey:   tx.data.ToFullShardKey,
		GasTokenID:       tx.data.GasTokenID,
		TransferTokenID:  tx.data.TransferTokenID,
	}

	return rlpHash(unsigntx)
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *EvmTransaction) WithSignature(sig []byte) (*EvmTransaction, error) {
	r, s, v, err := SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &EvmTransaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx *EvmTransaction, prv *ecdsa.PrivateKey) (*EvmTransaction, error) {
	h := tx.getUnsignedHash()
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(sig)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func SignatureValues(tx *EvmTransaction, sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = new(big.Int).SetBytes([]byte{sig[64] + 27})

	return R, S, V, nil
}
