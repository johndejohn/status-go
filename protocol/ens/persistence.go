package ens

import (
	"database/sql"
)

type ENSVerificationRecord struct {
	PublicKey           string
	Name                string
	Verified            bool
	VerifiedAt          uint64
	VerificationRetries uint64
}

type Persistence struct {
	db *sql.DB
}

func NewPersistence(db *sql.DB) *Persistence {
	return &Persistence{db: db}
}

func (p *Persistence) GetENSToBeVerified(now uint64) ([]*ENSVerificationRecord, error) {

	return nil, nil
}

func (p *Persistence) UpdateRecords(records []*ENSVerificationRecord) error {
	return nil
}
