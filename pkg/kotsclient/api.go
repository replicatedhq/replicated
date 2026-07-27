package kotsclient

import "context"

func (c *VendorV3Client) Get(ctx context.Context, path string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal(ctx, "GET", path, "")
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *VendorV3Client) Post(ctx context.Context, path string, body string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *VendorV3Client) Put(ctx context.Context, path string, body string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *VendorV3Client) Patch(ctx context.Context, path string, body string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal(ctx, "PATCH", path, body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
