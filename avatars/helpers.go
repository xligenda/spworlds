package avatars

import (
	"fmt"
	"net/url"
	"strconv"
)

func Size(pixels int) string {
	return strconv.Itoa(pixels)
}

func (c *Client) URL(uuid string, part Part, width int) string {
	u := fmt.Sprintf("%s/%s/%s", c.baseURL, part, uuid)
	if width > 0 {
		u += "?w=" + Size(width)
	}
	return u
}

func (c *Client) ParsedURL(uuid string, part Part, width int) (*url.URL, error) {
	return url.Parse(c.URL(uuid, part, width))
}
