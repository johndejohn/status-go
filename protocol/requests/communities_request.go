package requests

import (
	"errors"

	userimages "github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/images"
	"github.com/status-im/status-go/protocol/protobuf"
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
	if c.Name == "" || c.Description == "" || c.Membership == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
		return errors.New("CreateCommunity request invalid")
	}
	return nil
}

func (c *CreateCommunity) ToCommunityDescription() (*protobuf.CommunityDescription, error) {
	ci := &protobuf.ChatIdentity{
		DisplayName: c.Name,
		Description: c.Description,
	}

	if c.Image != "" {
		ciis := make(map[string]*protobuf.IdentityImage)
		imgs, err := userimages.GenerateIdentityImages(c.Image, 0, 0, 0, 0)
		if err != nil {
			return nil, err
		}
		for _, img := range imgs {
			ciis[img.Name] = adaptIdentityImageToProtobuf(img)
		}
		ci.Images = ciis
	}

	return &protobuf.CommunityDescription{
		Identity: ci,
		Permissions: &protobuf.CommunityPermissions{
			Access:  c.Membership,
			EnsOnly: c.EnsOnly,
		},
	}, nil
}
