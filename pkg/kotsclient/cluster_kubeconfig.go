package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

type GetClusterKubeconfigResponse struct {
	Kubeconfig []byte `json:"kubeconfig"`
	Error      string `json:"error"`
}

func (c *VendorV3Client) GetClusterKubeconfig(id string) ([]byte, error) {
	kubeconfig := GetClusterKubeconfigResponse{}

	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/cluster/%s/kubeconfig", id), http.StatusOK, nil, &kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubeconfig.Kubeconfig, nil
}
