package svg

import (
	"encoding/xml"
	"strings"
)

// appendStyleAttrs appends style attributes to the given attributes.
func appendStyleAttrs(attrs []xml.Attr, names ...string) ([]xml.Attr, error) {
	var style string

	for _, attr := range attrs {
		if attr.Name.Local == "style" {
			style = attr.Value
			break
		}
	}

	if style == "" {
		return attrs, nil
	}

	styleEl := strings.Split(style, ";")
	styleAttrs := make([]xml.Attr, 0, len(styleEl))
	for _, s := range styleEl {
		key, val, ok := strings.Cut(s, ":")
		if !ok {
			continue
		}

		key = strings.ToLower(strings.TrimSpace(key))

		if len(names) == 0 {
			styleAttrs = append(styleAttrs, xml.Attr{
				Name:  xml.Name{Local: key},
				Value: strings.TrimSpace(val),
			})
			continue
		}

		for _, name := range names {
			if key == name {
				attrs = append(attrs, xml.Attr{
					Name:  xml.Name{Local: key},
					Value: strings.TrimSpace(val),
				})
				break
			}
		}
	}

	return append(attrs, styleAttrs...), nil
}
