package shipclient

import (
	"reflect"
	"testing"

	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

func TestGraphQLClient_CreateCollector(t *testing.T) {
	type fields struct {
		GraphQLClient *graphql.Client
	}
	type args struct {
		appID string
		yaml  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.AppCollectorInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GraphQLClient{
				GraphQLClient: tt.fields.GraphQLClient,
			}
			got, err := c.CreateCollector(tt.args.appID, tt.args.yaml)
			if (err != nil) != tt.wantErr {
				t.Errorf("GraphQLClient.CreateCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GraphQLClient.CreateCollector() = %v, want %v", got, tt.want)
			}
		})
	}
}
