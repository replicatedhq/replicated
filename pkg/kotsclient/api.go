package kotsclient

func (c *VendorV3Client) Get(path string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal("GET", path, "")
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *VendorV3Client) Post(path string, body string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal("POST", path, body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *VendorV3Client) Put(path string, body string) ([]byte, error) {
	resp, err := c.DoJSONWithoutUnmarshal("PUT", path, body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
