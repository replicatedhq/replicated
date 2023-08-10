package kotsutil

import (
	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kotskinds/apis/kots/v1beta1"
	kotsscheme "github.com/replicatedhq/kotskinds/client/kotsclientset/scheme"
	releaseTypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	kotsscheme.AddToScheme(scheme.Scheme)
}

type OverlySimpleGVK struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

func GetKotsApplicationSpec(releaseSpecs []releaseTypes.KotsSingleSpec) (*kotsv1beta1.Application, error) {
	for _, r := range releaseSpecs {
		b := []byte(r.Content)
		o := OverlySimpleGVK{}

		if err := yaml.Unmarshal(b, &o); err != nil {
			// not a yaml file,
			continue
		}

		if o.APIVersion == "kots.io/v1beta1" && o.Kind == "Application" {
			decode := scheme.Codecs.UniversalDeserializer().Decode

			obj, gvk, err := decode(b, nil, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode content")
			}

			if gvk.String() != "kots.io/v1beta1, Kind=Application" {
				return nil, errors.Errorf("unexpected gvk: %s", gvk.String())
			}

			return obj.(*kotsv1beta1.Application), nil

		}
	}
	return nil, nil
}
