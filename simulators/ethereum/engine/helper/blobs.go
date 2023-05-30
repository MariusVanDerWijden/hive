package helper

import (
	"crypto/sha256"
	"errors"

	api "github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
)

func VersionedHashesFromBlobBundle(bb *api.BlobsBundle, commitmentVersion byte) ([]common.Hash, error) {
	if bb == nil {
		return nil, errors.New("nil blob bundle")
	}
	if bb.Commitments == nil {
		return nil, errors.New("nil commitments")
	}
	versionedHashes := make([]common.Hash, len(bb.Commitments))
	for i, commitment := range bb.Commitments {
		sha256Hash := sha256.Sum256(commitment[:])
		versionedHashes[i] = common.BytesToHash(append([]byte{commitmentVersion}, sha256Hash[1:]...))
	}
	return versionedHashes, nil
}
