package ens

import (
	"math"
	"strings"
)

type ENSVerificationRecord struct {
	PublicKey           string
	Name                string
	Clock               uint64
	Verified            bool
	VerifiedAt          uint64
	VerificationRetries uint64
	NextRetry           uint64
}

// We calculate if it's too early to retry, by exponentially backing off
func (e *ENSVerificationRecord) CalculateNextRetry() {
	e.NextRetry = e.VerifiedAt + ENSBackoffTimeSec*uint64(math.Exp2(float64(e.VerificationRetries)))
}

func (e *ENSVerificationRecord) Valid() bool {
	return e.Name != "" && strings.HasSuffix(e.Name, ".eth") && e.Clock > 0
}
