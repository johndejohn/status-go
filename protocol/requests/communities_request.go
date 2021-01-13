package requests

import (
	"errors"

	"github.com/ethereum/go-ethereum/log"
	userimages "github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/images"
	"github.com/status-im/status-go/protocol/protobuf"
)

var (
	ErrCreateCommunityInvalidName        = errors.New("create-community: invalid name")
	ErrCreateCommunityInvalidDescription = errors.New("create-community: invalid description")
	ErrCreateCommunityInvalidMembership  = errors.New("create-community: invalid membership")
)

type CreateCommunity struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Membership  protobuf.CommunityPermissions_Access
	EnsOnly     bool   `json:"ensOnly"`
	Image       string `json:"image"`
}

func adaptIdentityImageToProtobuf(img *userimages.IdentityImage) *protobuf.IdentityImage {
	return &protobuf.IdentityImage{
		Payload:    img.Payload,
		SourceType: protobuf.IdentityImage_RAW_PAYLOAD,
		ImageType:  images.ImageType(img.Payload),
	}
}

func (c *CreateCommunity) Validate() error {
	if c.Name == "" {
		return ErrCreateCommunityInvalidName
	}

	if c.Description == "" {
		return ErrCreateCommunityInvalidDescription
	}

	if c.Membership == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
		return ErrCreateCommunityInvalidMembership
	}

	return nil
}

func (c *CreateCommunity) ToCommunityDescription() (*protobuf.CommunityDescription, error) {
	ci := &protobuf.ChatIdentity{
		DisplayName: c.Name,
		Description: c.Description,
	}

	if c.Image != "" {
		log.Info("has-image", "image", c.Image)
		ciis := make(map[string]*protobuf.IdentityImage)
		imgs, err := userimages.GenerateIdentityImages(c.Image, 0, 0, 0, 0)
		if err != nil {
			return nil, err
		}
		for _, img := range imgs {
			ciis[img.Name] = adaptIdentityImageToProtobuf(img)
		}
		ci.Images = ciis
		log.Info("set images", "images", ci)
	}

	description := &protobuf.CommunityDescription{
		Identity: ci,
		Permissions: &protobuf.CommunityPermissions{
			Access:  c.Membership,
			EnsOnly: c.EnsOnly,
		},
	}
	return description, nil
}
