// # Test suite for blob tests
package suite_blobs

import (
	"math/big"
	"time"

	"github.com/ethereum/hive/simulators/ethereum/engine/client/hive_rpc"
	"github.com/ethereum/hive/simulators/ethereum/engine/helper"
	"github.com/ethereum/hive/simulators/ethereum/engine/test"
)

var (
	DATAHASH_START_ADDRESS = big.NewInt(0x100)
	DATAHASH_ADDRESS_COUNT = 1000

	// Fork specific constants
	DATA_GAS_PER_BLOB = uint64(0x20000)

	MIN_DATA_GASPRICE         = uint64(1)
	MAX_DATA_GAS_PER_BLOCK    = uint64(786432)
	TARGET_DATA_GAS_PER_BLOCK = uint64(393216)

	TARGET_BLOBS_PER_BLOCK = uint64(TARGET_DATA_GAS_PER_BLOCK / DATA_GAS_PER_BLOB)
	MAX_BLOBS_PER_BLOCK    = uint64(MAX_DATA_GAS_PER_BLOCK / DATA_GAS_PER_BLOB)

	DATA_GASPRICE_UPDATE_FRACTION = uint64(3338477)

	BLOB_COMMITMENT_VERSION_KZG = byte(0x01)
)

// Precalculate the first data gas cost increase
var (
	DATA_GAS_COST_INCREMENT_EXCEED_BLOBS = GetMinExcessDataBlobsForDataGasPrice(2)
)

func pUint64(v uint64) *uint64 {
	return &v
}

// Execution specification reference:
// https://github.com/ethereum/execution-apis/blob/main/src/engine/specification.md

// List of all blob tests
var Tests = []test.SpecInterface{
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Blob Transactions On Block 1, Cancun Genesis",
			About: `
			Tests the Cancun fork since genesis.
			`,
		},

		// We fork on genesis
		BlobsForkHeight: 0,

		BlobTestSequence: BlobTestSequence{
			// First, we send a couple of blob transactions on genesis,
			// with enough data gas cost to make sure they are included in the first block.
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},

			// We create the first payload, and verify that the blob transactions
			// are included in the payload.
			// We also verify that the blob transactions are included in the blobs bundle.
			NewPayloads{
				ExpectedIncludedBlobCount: TARGET_BLOBS_PER_BLOCK,
				ExpectedBlobs:             helper.GetBlobList(0, TARGET_BLOBS_PER_BLOCK),
			},

			// Try to increase the data gas cost of the blob transactions
			// by maxing out the number of blobs for the next payloads.
			SendBlobTransactions{
				BlobTransactionSendCount:      DATA_GAS_COST_INCREMENT_EXCEED_BLOBS/(MAX_BLOBS_PER_BLOCK-TARGET_BLOBS_PER_BLOCK) + 1,
				BlobsPerTransaction:           MAX_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},

			// Next payloads will have max data blobs each
			NewPayloads{
				PayloadCount:              DATA_GAS_COST_INCREMENT_EXCEED_BLOBS / (MAX_BLOBS_PER_BLOCK - TARGET_BLOBS_PER_BLOCK),
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK,
			},

			// But there will be an empty payload, since the data gas cost increased
			// and the last blob transaction was not included.
			NewPayloads{
				ExpectedIncludedBlobCount: 0,
			},

			// But it will be included in the next payload
			NewPayloads{
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK,
			},
		},
	},
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Blob Transaction Ordering, Single Account",
			About: `
			Send N blob transactions with MAX_BLOBS_PER_BLOCK-1 blobs each,
			using account A.
			Using same account, and an increased nonce from the previously sent
			transactions, send N blob transactions with 1 blob each.
			Verify that the payloads are created with the correct ordering:
			 - The first payloads must include the first N blob transactions.
			 - The last payloads must include the last single-blob transactions.
			All transactions have sufficient data gas price to be included any
			of the payloads.
			`,
		},

		// We fork on genesis
		BlobsForkHeight: 0,

		BlobTestSequence: BlobTestSequence{
			// First send the MAX_BLOBS_PER_BLOCK-1 blob transactions.
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           MAX_BLOBS_PER_BLOCK - 1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
			},
			// Then send the single-blob transactions
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
			},

			// First four payloads have MAX_BLOBS_PER_BLOCK-1 blobs each
			NewPayloads{
				PayloadCount:              4,
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK - 1,
			},

			// The rest of the payloads have full blobs
			NewPayloads{
				PayloadCount:              2,
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK,
			},
		},
	},

	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Blob Transaction Ordering, Multiple Accounts",
			About: `
			Send N blob transactions with MAX_BLOBS_PER_BLOCK-1 blobs each,
			using account A.
			Send N blob transactions with 1 blob each from account B.
			Verify that the payloads are created with the correct ordering:
			 - All payloads must have full blobs.
			All transactions have sufficient data gas price to be included any
			of the payloads.
			`,
		},

		// We fork on genesis
		BlobsForkHeight: 0,

		BlobTestSequence: BlobTestSequence{
			// First send the MAX_BLOBS_PER_BLOCK-1 blob transactions from
			// account A.
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           MAX_BLOBS_PER_BLOCK - 1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
				AccountIndex:                  0,
			},
			// Then send the single-blob transactions from account B
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
				AccountIndex:                  1,
			},

			// All payloads have full blobs
			NewPayloads{
				PayloadCount:              5,
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK,
			},
		},
	},

	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Blob Transaction Ordering, Multiple Clients",
			About: `
			Send N blob transactions with MAX_BLOBS_PER_BLOCK-1 blobs each,
			using account A, to client A.
			Send N blob transactions with 1 blob each from account B, to client
			B.
			Verify that the payloads are created with the correct ordering:
			 - All payloads must have full blobs.
			All transactions have sufficient data gas price to be included any
			of the payloads.
			`,
		},

		// We fork on genesis
		BlobsForkHeight: 0,

		BlobTestSequence: BlobTestSequence{
			// Start a secondary client to also receive blob transactions
			LaunchClients{
				EngineStarter: hive_rpc.HiveRPCEngineStarter{},
				// Skip adding the second client to the CL Mock to guarantee
				// that all payloads are produced by client A.
				// This is done to not have client B prioritizing single-blob
				// transactions to fill one single payload.
				SkipAddingToCLMock: true,
			},

			// Create a block without any blobs to get past genesis
			NewPayloads{
				PayloadCount:              1,
				ExpectedIncludedBlobCount: 0,
			},

			// First send the MAX_BLOBS_PER_BLOCK-1 blob transactions from
			// account A, to client A.
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           MAX_BLOBS_PER_BLOCK - 1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
				AccountIndex:                  0,
				ClientIndex:                   0,
			},
			// Then send the single-blob transactions from account B, to client
			// B.
			SendBlobTransactions{
				BlobTransactionSendCount:      5,
				BlobsPerTransaction:           1,
				BlobTransactionMaxDataGasCost: big.NewInt(100),
				AccountIndex:                  1,
				ClientIndex:                   1,
			},

			// All payloads have full blobs
			NewPayloads{
				PayloadCount:              5,
				ExpectedIncludedBlobCount: MAX_BLOBS_PER_BLOCK,
				// Wait a bit more on before requesting the built payload from the client
				GetPayloadDelay: 2,
			},
		},
	},

	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Replace Blob Transactions",
			About: `
			Test sending multiple blob transactions with the same nonce, but
			higher gas tip so the transaction is replaced.
			`,
		},

		// We fork on genesis
		BlobsForkHeight: 0,

		BlobTestSequence: BlobTestSequence{
			// Send multiple blob transactions with the same nonce.
			SendBlobTransactions{ // Blob ID 0
				BlobTransactionSendCount:      1,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
				BlobTransactionGasFeeCap:      big.NewInt(1e9),
				BlobTransactionGasTipCap:      big.NewInt(1e9),
			},
			SendBlobTransactions{ // Blob ID 1
				BlobTransactionSendCount:      1,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
				BlobTransactionGasFeeCap:      big.NewInt(1e10),
				BlobTransactionGasTipCap:      big.NewInt(1e10),
				ReplaceTransactions:           true,
			},
			SendBlobTransactions{ // Blob ID 2
				BlobTransactionSendCount:      1,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
				BlobTransactionGasFeeCap:      big.NewInt(1e11),
				BlobTransactionGasTipCap:      big.NewInt(1e11),
				ReplaceTransactions:           true,
			},
			SendBlobTransactions{ // Blob ID 3
				BlobTransactionSendCount:      1,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
				BlobTransactionGasFeeCap:      big.NewInt(1e12),
				BlobTransactionGasTipCap:      big.NewInt(1e12),
				ReplaceTransactions:           true,
			},

			// We create the first payload, which must contain the blob tx
			// with the higher tip.
			NewPayloads{
				ExpectedIncludedBlobCount: 1,
				ExpectedBlobs:             []helper.BlobID{3},
			},
		},
	},
	// Test versioned hashes in Engine API NewPayloadV3
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Missing Hash",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is missing one of the hashes.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0},
				},
			},
		},
	},
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Extra Hash",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is has an extra hash for a blob that is not in the payload.
			`,
		},
		// TODO: It could be worth it to also test this with a blob that is in the
		// mempool but was not included in the payload.
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1, 2},
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Out of Order",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is out of order.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{1, 0},
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Repeated Hash",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a blob that is repeated in the array.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1, 1},
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Incorrect Hash",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a blob that is repeated in the array.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 2},
				},
			},
		},
	},
	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Incorrect Version",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a single blob that has an incorrect version.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs:        []helper.BlobID{0, 1},
					HashVersions: []byte{BLOB_COMMITMENT_VERSION_KZG, BLOB_COMMITMENT_VERSION_KZG + 1},
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Nil Hashes",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is nil, even though the fork has already happened.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: nil,
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Empty Hashes",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is empty, even though there are blobs in the payload.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{},
				},
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Non-Empty Hashes",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is contains hashes, even though there are no blobs in the payload.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{
				ExpectedIncludedBlobCount: 0,
				ExpectedBlobs:             []helper.BlobID{},
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1},
				},
			},
		},
	},

	// Test versioned hashes in Engine API NewPayloadV3 on syncing clients
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Missing Hash (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is missing one of the hashes.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Extra Hash (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is has an extra hash for a blob that is not in the payload.
			`,
		},
		// TODO: It could be worth it to also test this with a blob that is in the
		// mempool but was not included in the payload.
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1, 2},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Out of Order (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is out of order.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},
			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{1, 0},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Repeated Hash (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a blob that is repeated in the array.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1, 1},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Incorrect Hash (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a blob that is repeated in the array.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 2},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},
	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Incorrect Version (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			has a single blob that has an incorrect version.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs:        []helper.BlobID{0, 1},
					HashVersions: []byte{BLOB_COMMITMENT_VERSION_KZG, BLOB_COMMITMENT_VERSION_KZG + 1},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Nil Hashes (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is nil, even though the fork has already happened.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: nil,
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Empty Hashes (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is empty, even though there are blobs in the payload.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			SendBlobTransactions{
				BlobTransactionSendCount:      TARGET_BLOBS_PER_BLOCK,
				BlobTransactionMaxDataGasCost: big.NewInt(1),
			},
			NewPayloads{
				ExpectedIncludedBlobCount: 2,
				ExpectedBlobs:             []helper.BlobID{0, 1},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	&BlobsBaseSpec{
		Spec: test.Spec{
			Name: "NewPayloadV3 Versioned Hashes, Non-Empty Hashes (Syncing)",
			About: `
			Tests VersionedHashes in Engine API NewPayloadV3 where the array
			is contains hashes, even though there are no blobs in the payload.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{}, // Send new payload so the parent is unknown to the secondary client
			NewPayloads{
				ExpectedIncludedBlobCount: 0,
				ExpectedBlobs:             []helper.BlobID{},
			},

			LaunchClients{
				EngineStarter:            hive_rpc.HiveRPCEngineStarter{},
				SkipAddingToCLMock:       true,
				SkipConnectingToBootnode: true, // So the client is in a perpetual syncing state
			},
			SendModifiedLatestPayload{
				ClientID: 1,
				VersionedHashes: &VersionedHashes{
					Blobs: []helper.BlobID{0, 1},
				},
				ExpectedStatus: test.Invalid,
			},
		},
	},

	// DataGasUsed, ExcessDataGas Negative Tests
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Incorrect DataGasUsed: Non-Zero on Zero Blobs",
			About: `
			Send a payload with zero blobs, but non-zero DataGasUsed.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{
				ExpectedIncludedBlobCount: 0,
				PayloadCustomizer: &helper.CustomPayloadData{
					DataGasUsed: pUint64(1),
				},
			},
		},
	},
	&BlobsBaseSpec{

		Spec: test.Spec{
			Name: "Incorrect DataGasUsed: DATA_GAS_PER_BLOB on Zero Blobs",
			About: `
			Send a payload with zero blobs, but non-zero DataGasUsed.
			`,
		},
		BlobTestSequence: BlobTestSequence{
			NewPayloads{
				ExpectedIncludedBlobCount: 0,
				PayloadCustomizer: &helper.CustomPayloadData{
					DataGasUsed: pUint64(DATA_GAS_PER_BLOB),
				},
			},
		},
	},
}

// Blobs base spec
// This struct contains the base spec for all blob tests. It contains the
// timestamp increments per block, the withdrawals fork height, and the list of
// payloads to produce during the test.
type BlobsBaseSpec struct {
	test.Spec
	TimeIncrements  uint64 // Timestamp increments per block throughout the test
	GetPayloadDelay uint64 // Delay between FcU and GetPayload calls
	BlobsForkHeight uint64 // Withdrawals activation fork height
	BlobTestSequence
}

// Base test case execution procedure for blobs tests.
func (bs *BlobsBaseSpec) Execute(t *test.Env) {

	t.CLMock.WaitForTTD()

	blobTestCtx := &BlobTestContext{
		Env:            t,
		TestBlobTxPool: new(TestBlobTxPool),
	}

	if bs.GetPayloadDelay != 0 {
		t.CLMock.PayloadProductionClientDelay = time.Duration(bs.GetPayloadDelay) * time.Second
	}

	for stepId, step := range bs.BlobTestSequence {
		t.Logf("INFO: Executing step %d: %s", stepId+1, step.Description())
		if err := step.Execute(blobTestCtx); err != nil {
			t.Fatalf("FAIL: Error executing step %d: %v", stepId+1, err)
		}
	}

}
