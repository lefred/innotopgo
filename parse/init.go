package parse

import (
	"fmt"
	"net/url"
	"strings"
)

func Parse(uri_to_parse string) (string, error) {

	uri, err := url.Parse(uri_to_parse)
	if err != nil {
		return "", err
	}

	if uri.Scheme == "" || !strings.HasPrefix(uri_to_parse, "mysql://") {
		uri_to_parse = fmt.Sprintf("mysql://%s", uri_to_parse)
		uri, err = url.Parse(uri_to_parse)
		if err != nil {
			return "", err
		}
	}
	pwd, _ := uri.User.Password()

	mysql_uri := fmt.Sprintf("%s:%s@tcp(%s)/", uri.User.Username(), pwd, uri.Host)

	return mysql_uri, nil
}
