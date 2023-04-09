package utils

import (
	"fmt"
	"strings"

	"github.com/rs/xid"
)

func JoinID(namespace string, id xid.ID) string {
	if strings.ContainsRune(namespace, ':') {
		panic(fmt.Sprintf("namespace %s contains ':' rune", namespace))
	}

	return fmt.Sprintf("%s:%s", namespace, id.String())
}

func SplitID(id string) (string, xid.ID, error) {
	splitted := strings.SplitN(id, ":", 2)
	if len(splitted) != 2 {
		return "", xid.ID{}, fmt.Errorf("invalid id: %s", id)
	}

	idOut, err := xid.FromString(splitted[1])
	if err != nil {
		return "", xid.ID{}, err
	}

	return splitted[0], idOut, nil
}
