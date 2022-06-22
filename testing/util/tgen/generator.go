package tgen

import (
	"math/rand"

	"github.com/prysmaticlabs/go-bitfield"
	types "github.com/prysmaticlabs/prysm/consensus-types/primitives"
	enginev1 "github.com/prysmaticlabs/prysm/proto/engine/v1"
	eth "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

var b20 = make([]byte, 20)
var b32 = make([]byte, 32)
var b48 = make([]byte, 48)
var b64 = make([]byte, 64)
var b96 = make([]byte, 96)
var b256 = make([]byte, 256)

func init() {
	_, err := rand.Read(b20)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 20; i++ {
		if b20[i] == 0x00 {
			b20[i] = uint8(rand.Int())
		}
	}
	_, err = rand.Read(b32)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 32; i++ {
		if b32[i] == 0x00 {
			b32[i] = uint8(rand.Int())
		}
	}
	_, err = rand.Read(b48)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 48; i++ {
		if b48[i] == 0x00 {
			b48[i] = uint8(rand.Int())
		}
	}
	_, err = rand.Read(b64)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 64; i++ {
		if b64[i] == 0x00 {
			b64[i] = uint8(rand.Int())
		}
	}
	_, err = rand.Read(b96)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 96; i++ {
		if b96[i] == 0x00 {
			b96[i] = uint8(rand.Int())
		}
	}
	_, err = rand.Read(b256)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 256; i++ {
		if b256[i] == 0x00 {
			b256[i] = uint8(rand.Int())
		}
	}
}

type byteSlices struct {
	B20  []byte
	B32  []byte
	B48  []byte
	B64  []byte
	B96  []byte
	B256 []byte
}

// BlockFields is a collection of all fields that make up a block
type BlockFields struct {
	byteSlices
	Slot              types.Slot
	ProposerIndex     types.ValidatorIndex
	Deposits          []*eth.Deposit
	Atts              []*eth.Attestation
	ProposerSlashings []*eth.ProposerSlashing
	AttesterSlashings []*eth.AttesterSlashing
	VoluntaryExits    []*eth.SignedVoluntaryExit
	SyncAggregate     *eth.SyncAggregate
	ExecPayload       *enginev1.ExecutionPayload
	ExecPayloadHeader *eth.ExecutionPayloadHeader
}

// GetBlockFields returns an instance of BlockFields with generated values
func GetBlockFields() BlockFields {
	deposits := make([]*eth.Deposit, 16)
	for i := range deposits {
		deposits[i] = &eth.Deposit{}
		deposits[i].Proof = make([][]byte, 33)
		for j := range deposits[i].Proof {
			deposits[i].Proof[j] = b32
		}
		deposits[i].Data = &eth.Deposit_Data{
			PublicKey:             b48,
			WithdrawalCredentials: b32,
			Amount:                128,
			Signature:             b96,
		}
	}
	atts := make([]*eth.Attestation, 128)
	for i := range atts {
		atts[i] = &eth.Attestation{}
		atts[i].Signature = b96
		atts[i].AggregationBits = bitfield.NewBitlist(1)
		atts[i].Data = &eth.AttestationData{
			Slot:            128,
			CommitteeIndex:  128,
			BeaconBlockRoot: b32,
			Source: &eth.Checkpoint{
				Epoch: 128,
				Root:  b32,
			},
			Target: &eth.Checkpoint{
				Epoch: 128,
				Root:  b32,
			},
		}
	}
	proposerSlashing := &eth.ProposerSlashing{
		Header_1: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          128,
				ProposerIndex: 128,
				ParentRoot:    b32,
				StateRoot:     b32,
				BodyRoot:      b32,
			},
			Signature: b96,
		},
		Header_2: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          128,
				ProposerIndex: 128,
				ParentRoot:    b32,
				StateRoot:     b32,
				BodyRoot:      b32,
			},
			Signature: b96,
		},
	}
	attesterSlashing := &eth.AttesterSlashing{
		Attestation_1: &eth.IndexedAttestation{
			AttestingIndices: []uint64{1, 2, 8},
			Data: &eth.AttestationData{
				Slot:            128,
				CommitteeIndex:  128,
				BeaconBlockRoot: b32,
				Source: &eth.Checkpoint{
					Epoch: 128,
					Root:  b32,
				},
				Target: &eth.Checkpoint{
					Epoch: 128,
					Root:  b32,
				},
			},
			Signature: b96,
		},
		Attestation_2: &eth.IndexedAttestation{
			AttestingIndices: []uint64{1, 2, 8},
			Data: &eth.AttestationData{
				Slot:            128,
				CommitteeIndex:  128,
				BeaconBlockRoot: b32,
				Source: &eth.Checkpoint{
					Epoch: 128,
					Root:  b32,
				},
				Target: &eth.Checkpoint{
					Epoch: 128,
					Root:  b32,
				},
			},
			Signature: b96,
		},
	}
	voluntaryExit := &eth.SignedVoluntaryExit{
		Exit: &eth.VoluntaryExit{
			Epoch:          128,
			ValidatorIndex: 128,
		},
		Signature: b96,
	}
	syncAggregate := &eth.SyncAggregate{
		SyncCommitteeBits:      b64,
		SyncCommitteeSignature: b96,
	}
	execPayload := &enginev1.ExecutionPayload{
		ParentHash:    b32,
		FeeRecipient:  b20,
		StateRoot:     b32,
		ReceiptsRoot:  b32,
		LogsBloom:     b256,
		PrevRandao:    b32,
		BlockNumber:   128,
		GasLimit:      128,
		GasUsed:       128,
		Timestamp:     128,
		ExtraData:     b32,
		BaseFeePerGas: b32,
		BlockHash:     b32,
		Transactions: [][]byte{
			[]byte("transaction1"),
			[]byte("transaction2"),
			[]byte("transaction8"),
		},
	}
	execPayloadHeader := &eth.ExecutionPayloadHeader{
		ParentHash:       b32,
		FeeRecipient:     b20,
		StateRoot:        b32,
		ReceiptsRoot:     b32,
		LogsBloom:        b256,
		PrevRandao:       b32,
		BlockNumber:      128,
		GasLimit:         128,
		GasUsed:          128,
		Timestamp:        128,
		ExtraData:        b32,
		BaseFeePerGas:    b32,
		BlockHash:        b32,
		TransactionsRoot: b32,
	}

	return BlockFields{
		byteSlices: byteSlices{
			B20:  b20,
			B32:  b32,
			B48:  b48,
			B96:  b96,
			B256: b256,
		},
		Slot:              128,
		ProposerIndex:     128,
		Deposits:          deposits,
		Atts:              atts,
		ProposerSlashings: []*eth.ProposerSlashing{proposerSlashing},
		AttesterSlashings: []*eth.AttesterSlashing{attesterSlashing},
		VoluntaryExits:    []*eth.SignedVoluntaryExit{voluntaryExit},
		SyncAggregate:     syncAggregate,
		ExecPayload:       execPayload,
		ExecPayloadHeader: execPayloadHeader,
	}
}

// PbSignedBeaconBlockPhase0 generates a SignedBeaconBlock
func PbSignedBeaconBlockPhase0() *eth.SignedBeaconBlock {
	f := GetBlockFields()
	return &eth.SignedBeaconBlock{
		Block:     PbBeaconBlockPhase0(),
		Signature: f.B96,
	}
}

// PbSignedBeaconBlockAltair generates a SignedBeaconBlockAltair
func PbSignedBeaconBlockAltair() *eth.SignedBeaconBlockAltair {
	f := GetBlockFields()
	return &eth.SignedBeaconBlockAltair{
		Block:     PbBeaconBlockAltair(),
		Signature: f.B96,
	}
}

// PbSignedBeaconBlockBellatrix generates a SignedBeaconBlockBellatrix
func PbSignedBeaconBlockBellatrix() *eth.SignedBeaconBlockBellatrix {
	f := GetBlockFields()
	return &eth.SignedBeaconBlockBellatrix{
		Block:     PbBeaconBlockBellatrix(),
		Signature: f.B96,
	}
}

// PbSignedBlindedBeaconBlockBellatrix generates a SignedBlindedBeaconBlockBellatrix
func PbSignedBlindedBeaconBlockBellatrix() *eth.SignedBlindedBeaconBlockBellatrix {
	f := GetBlockFields()
	return &eth.SignedBlindedBeaconBlockBellatrix{
		Block:     PbBlindedBeaconBlockBellatrix(),
		Signature: f.B96,
	}
}

// PbBeaconBlockPhase0 generates a BeaconBlock
func PbBeaconBlockPhase0() *eth.BeaconBlock {
	f := GetBlockFields()
	return &eth.BeaconBlock{
		Slot:          128,
		ProposerIndex: 128,
		ParentRoot:    f.B32,
		StateRoot:     f.B32,
		Body:          PbBeaconBlockBodyPhase0(),
	}
}

// PbBeaconBlockAltair generates a BeaconBlockAltair
func PbBeaconBlockAltair() *eth.BeaconBlockAltair {
	f := GetBlockFields()
	return &eth.BeaconBlockAltair{
		Slot:          128,
		ProposerIndex: 128,
		ParentRoot:    f.B32,
		StateRoot:     f.B32,
		Body:          PbBeaconBlockBodyAltair(),
	}
}

// PbBeaconBlockBellatrix generates a BeaconBlockBellatrix
func PbBeaconBlockBellatrix() *eth.BeaconBlockBellatrix {
	f := GetBlockFields()
	return &eth.BeaconBlockBellatrix{
		Slot:          128,
		ProposerIndex: 128,
		ParentRoot:    f.B32,
		StateRoot:     f.B32,
		Body:          PbBeaconBlockBodyBellatrix(),
	}
}

// PbBlindedBeaconBlockBellatrix generates a BlindedBeaconBlockBellatrix
func PbBlindedBeaconBlockBellatrix() *eth.BlindedBeaconBlockBellatrix {
	f := GetBlockFields()
	return &eth.BlindedBeaconBlockBellatrix{
		Slot:          128,
		ProposerIndex: 128,
		ParentRoot:    f.B32,
		StateRoot:     f.B32,
		Body:          PbBlindedBeaconBlockBodyBellatrix(),
	}
}

// PbBeaconBlockBodyPhase0 generates a BeaconBlockBody
func PbBeaconBlockBodyPhase0() *eth.BeaconBlockBody {
	f := GetBlockFields()
	return &eth.BeaconBlockBody{
		RandaoReveal: f.B96,
		Eth1Data: &eth.Eth1Data{
			DepositRoot:  f.B32,
			DepositCount: 128,
			BlockHash:    f.B32,
		},
		Graffiti:          f.B32,
		ProposerSlashings: f.ProposerSlashings,
		AttesterSlashings: f.AttesterSlashings,
		Attestations:      f.Atts,
		Deposits:          f.Deposits,
		VoluntaryExits:    f.VoluntaryExits,
	}
}

// PbBeaconBlockBodyAltair generates a BeaconBlockBodyAltair
func PbBeaconBlockBodyAltair() *eth.BeaconBlockBodyAltair {
	f := GetBlockFields()
	return &eth.BeaconBlockBodyAltair{
		RandaoReveal: f.B96,
		Eth1Data: &eth.Eth1Data{
			DepositRoot:  f.B32,
			DepositCount: 128,
			BlockHash:    f.B32,
		},
		Graffiti:          f.B32,
		ProposerSlashings: f.ProposerSlashings,
		AttesterSlashings: f.AttesterSlashings,
		Attestations:      f.Atts,
		Deposits:          f.Deposits,
		VoluntaryExits:    f.VoluntaryExits,
		SyncAggregate:     f.SyncAggregate,
	}
}

// PbBeaconBlockBodyBellatrix generates a BeaconBlockBodyBellatrix
func PbBeaconBlockBodyBellatrix() *eth.BeaconBlockBodyBellatrix {
	f := GetBlockFields()
	return &eth.BeaconBlockBodyBellatrix{
		RandaoReveal: f.B96,
		Eth1Data: &eth.Eth1Data{
			DepositRoot:  f.B32,
			DepositCount: 128,
			BlockHash:    f.B32,
		},
		Graffiti:          f.B32,
		ProposerSlashings: f.ProposerSlashings,
		AttesterSlashings: f.AttesterSlashings,
		Attestations:      f.Atts,
		Deposits:          f.Deposits,
		VoluntaryExits:    f.VoluntaryExits,
		SyncAggregate:     f.SyncAggregate,
		ExecutionPayload:  f.ExecPayload,
	}
}

// PbBlindedBeaconBlockBodyBellatrix generates a BlindedBeaconBlockBodyBellatrix
func PbBlindedBeaconBlockBodyBellatrix() *eth.BlindedBeaconBlockBodyBellatrix {
	f := GetBlockFields()
	return &eth.BlindedBeaconBlockBodyBellatrix{
		RandaoReveal: f.B96,
		Eth1Data: &eth.Eth1Data{
			DepositRoot:  f.B32,
			DepositCount: 128,
			BlockHash:    f.B32,
		},
		Graffiti:               f.B32,
		ProposerSlashings:      f.ProposerSlashings,
		AttesterSlashings:      f.AttesterSlashings,
		Attestations:           f.Atts,
		Deposits:               f.Deposits,
		VoluntaryExits:         f.VoluntaryExits,
		SyncAggregate:          f.SyncAggregate,
		ExecutionPayloadHeader: f.ExecPayloadHeader,
	}
}
