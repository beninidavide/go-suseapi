package bugzilla

func (c *Client) PatchBug(source []byte) []byte {
	return c.patchBug(source)
}
