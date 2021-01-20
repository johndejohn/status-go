package communities

import (
	"crypto/ecdsa"
	"database/sql"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/protobuf"
)

type Persistence struct {
	db     *sql.DB
	logger *zap.Logger
}

func (p *Persistence) SaveCommunity(community *Community) error {
	id := community.ID()
	privateKey := community.PrivateKey()
	description, err := community.ToBytes()
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`INSERT INTO communities_communities (id, private_key, description, joined, verified) VALUES (?, ?, ?,?,?)`, id, crypto.FromECDSA(privateKey), description, community.config.Joined, community.config.Verified)
	return err
}

func (p *Persistence) queryCommunities(memberIdentity *ecdsa.PublicKey, query string) (response []*Community, err error) {

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			// Don't shadow original error
			_ = rows.Close()
			return

		}
		err = rows.Close()
	}()

	for rows.Next() {
		var publicKeyBytes, privateKeyBytes, descriptionBytes []byte
		var joined bool
		var verified bool
		err := rows.Scan(&publicKeyBytes, &privateKeyBytes, &descriptionBytes, &joined, &verified)
		if err != nil {
			return nil, err
		}

		org, err := unmarshalCommunityFromDB(memberIdentity, publicKeyBytes, privateKeyBytes, descriptionBytes, joined, verified, p.logger)
		if err != nil {
			return nil, err
		}
		response = append(response, org)
	}

	return response, nil

}

func (p *Persistence) AllCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := `SELECT id, private_key, description,joined,verified FROM communities_communities`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) JoinedCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := `SELECT id, private_key, description,joined,verified FROM communities_communities WHERE joined`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) CreatedCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := `SELECT id, private_key, description,joined,verified FROM communities_communities WHERE private_key IS NOT NULL`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) GetByID(memberIdentity *ecdsa.PublicKey, id []byte) (*Community, error) {
	var publicKeyBytes, privateKeyBytes, descriptionBytes []byte
	var joined bool
	var verified bool

	err := p.db.QueryRow(`SELECT id, private_key, description, joined,verified FROM communities_communities WHERE id = ?`, id).Scan(&publicKeyBytes, &privateKeyBytes, &descriptionBytes, &joined, &verified)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return unmarshalCommunityFromDB(memberIdentity, publicKeyBytes, privateKeyBytes, descriptionBytes, joined, verified, p.logger)
}

func unmarshalCommunityFromDB(memberIdentity *ecdsa.PublicKey, publicKeyBytes, privateKeyBytes, descriptionBytes []byte, joined, verified bool, logger *zap.Logger) (*Community, error) {

	var privateKey *ecdsa.PrivateKey
	var err error

	if privateKeyBytes != nil {
		privateKey, err = crypto.ToECDSA(privateKeyBytes)
		if err != nil {
			return nil, err
		}
	}
	metadata := &protobuf.ApplicationMetadataMessage{}

	err = proto.Unmarshal(descriptionBytes, metadata)
	if err != nil {
		return nil, err
	}

	description := &protobuf.CommunityDescription{}

	err = proto.Unmarshal(metadata.Payload, description)
	if err != nil {
		return nil, err
	}

	id, err := crypto.DecompressPubkey(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	config := Config{
		PrivateKey:                    privateKey,
		CommunityDescription:          description,
		MemberIdentity:                memberIdentity,
		MarshaledCommunityDescription: descriptionBytes,
		Logger:                        logger,
		ID:                            id,
		Verified:                      verified,
		Joined:                        joined,
	}
	return New(config)
}

func (p *Persistence) SaveRequestToJoin(request *RequestToJoin) error {
	_, err := p.db.Exec(`INSERT INTO communities_requests_to_join(id,public_key,clock,ens_name,chat_id,community_id,state) VALUES (?, ?, ?, ?, ?, ?, ?)`, request.ID, request.PublicKey, request.Clock, request.ENSName, request.ChatID, request.CommunityID, request.State)
	return err
}

func (p *Persistence) PendingRequestsToJoinForUser(pk string) ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ? AND public_key = ?`, RequestToJoinStatePending, pk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) PendingRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ? AND community_id = ?`, RequestToJoinStatePending, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) SetRequestToJoinState(communityID []byte, state uint) error {
	_, err := p.db.Exec(`UPDATE communities_requests_to_join SET state = ? WHERE community_id = ?`, state, communityID)
	return err
}

func (p *Persistence) GetRequestToJoin(id []byte) (*RequestToJoin, error) {
	request := &RequestToJoin{}
	err := p.db.QueryRow(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE id = ?`, id).Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
	if err != nil {
		return nil, err
	}

	return request, nil
}
