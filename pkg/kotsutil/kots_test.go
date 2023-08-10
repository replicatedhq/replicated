package kotsutil

import (
	"reflect"
	"testing"

	kotsv1beta1 "github.com/replicatedhq/kotskinds/apis/kots/v1beta1"
	releaseTypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetKotsApplicationSpec(t *testing.T) {
	tests := []struct {
		name         string
		releaseSpecs []releaseTypes.KotsSingleSpec
		want         *kotsv1beta1.Application
		wantErr      bool
	}{
		{
			name:         "no release specs",
			releaseSpecs: []releaseTypes.KotsSingleSpec{},
			want:         nil,
			wantErr:      false,
		},
		{
			name: "no yamls",
			releaseSpecs: []releaseTypes.KotsSingleSpec{
				{
					Content: "text content",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "one yaml, not an application",
			releaseSpecs: []releaseTypes.KotsSingleSpec{
				{
					Content: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: foo",
				},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "one yaml, application",
			releaseSpecs: []releaseTypes.KotsSingleSpec{
				{
					Content: "apiVersion: kots.io/v1beta1\nkind: Application\nmetadata:\n  name: foo",
				},
			},
			want: &kotsv1beta1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kots.io/v1beta1",
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		}, {
			name: "two yamls, application",
			releaseSpecs: []releaseTypes.KotsSingleSpec{
				{
					Content: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: foo",
				},
				{
					Content: "apiVersion: kots.io/v1beta1\nkind: Application\nmetadata:\n  name: foo",
				},
			},
			want: &kotsv1beta1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kots.io/v1beta1",
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKotsApplicationSpec(tt.releaseSpecs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKotsApplicationSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKotsApplicationSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
