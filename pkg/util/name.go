package util

import "github.com/moby/moby/pkg/namesgenerator"

func GenerateName() string {
	return namesgenerator.GetRandomName(0)
}
