package bugzilla

func (c *Client) PatchBug(source string) string {
	return c.patchBug(source)
}
