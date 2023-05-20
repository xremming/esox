package models

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/xid"
)

type ID struct{ xid.ID }

func NewID() ID {
	return ID{xid.New()}
}

func IDFromString(v string) (ID, error) {
	id, err := xid.FromString(v)
	if err != nil {
		return ID{}, err
	}

	return ID{id}, nil
}

func (id *ID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: id.String()}, nil
}

func (id *ID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	in, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return errors.New("id can only be unmarshalled from a string")
	}

	v, err := xid.FromString(in.Value)
	if err != nil {
		return err
	}

	id.ID = v

	return nil
}

// NSID is a namespaced ID.
type NSID struct {
	Namespace string
	ID        xid.ID
}

func NewNSID(namespace string) NSID {
	return NSID{
		Namespace: namespace,
		ID:        xid.New(),
	}
}

func (nsid NSID) String() string {
	return fmt.Sprintf("%s:%s", nsid.Namespace, nsid.ID.String())
}

func (nsid *NSID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: nsid.String()}, nil
}

func splitID(s string) (NSID, error) {
	namespace, idString := cutRight(s, ':')

	id, err := xid.FromString(idString)
	if err != nil {
		return NSID{}, err
	}

	return NSID{
		Namespace: namespace,
		ID:        id,
	}, nil
}

func (nsid *NSID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	in, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return errors.New("nsid can only be unmarshalled from a string")
	}

	var err error
	v, err := splitID(in.Value)
	if err != nil {
		return err
	}

	*nsid = v

	return nil
}
