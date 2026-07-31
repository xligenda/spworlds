package avatars

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func Size(pixels int) string {
	return strconv.Itoa(pixels)
}

func (c *Client) URL(player string, part Part, width int) string {
	if width < 0 {
		width = SizeDefault
	}

	return fmt.Sprintf("%s/%s/%s?width=%s",
		strings.TrimRight(c.baseURL, "/"),
		part,
		url.PathEscape(player),
		Size(width),
	)
}

func (c *Client) ParsedURL(uuid string, part Part, width int) (*url.URL, error) {
	return url.Parse(c.URL(uuid, part, width))
}
