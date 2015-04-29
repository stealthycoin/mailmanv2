package collector

import (
	"strings"
	"errors"
	"bytes"
	"log"
)

//
// Figure some shit out about a template string
//
func lex(str string) ([]int, error) {
	last := ""
	idxs := make([]int, 0, 0)
	for i, r := range str {
		if last != `\` &&  string(r) == "%" {
			idxs = append(idxs, i)
		}
		last = string(r)
	}

	if len(idxs) % 2 != 0 {
		// Odd number of % invalid template
		return nil, errors.New("Odd number of % in first template: " + str)
	}

	return idxs, nil
}


//
// Extracts field name and content from a template
//
func extract(field string) (string, string) {
	field = strings.Trim(field, " %")// get rid of whitespace and trailing %
	i := strings.Index(field, " ")
	if i == -1 {
		// No content just a field name
		return field, ""
	}

	// First word is the fieldname, rest is content
	return field[:i], field[i+1:]
}

//
// Merges two templates together
//
func tmpl_merge(a, b string) (string, error) {
	// 	Collect % indicies
	a_idxs, err := lex(a)
	if err != nil {
		log.Println(err)
	}
	b_idxs, err := lex(b)
	if err != nil {
		log.Println(err)
	}

	// No templates here just return the second string
	if len(b_idxs) == 0 && len(a_idxs) == 0 {
		return b, nil
	}

	a_map := make(map[string]string)

	// Map content for child template
	for i, v := range a_idxs {
		if i % 2 != 0 {
			key, value := extract(a[a_idxs[i-1]:v])
			a_map[key] = value
		}
	}

	// Result string
	var buffer bytes.Buffer

	// Merge map into parent template
	for i, v := range b_idxs {
		if i % 2 != 0 {
			// Odd percent tag is a closing one
			key, value := extract(b[b_idxs[i-1]:v])
			if a_v, ok := a_map[key]; ok {
				// Child template has the same key, merge
				buffer.WriteString("%")
				buffer.WriteString(key)
				buffer.WriteString(" ")
				buffer.WriteString(value)
				buffer.WriteString(", ")
				buffer.WriteString(a_v)
			}
		} else {
			// Even template tag is an opening one, copy from last
			// tag to this one
			if i > 0 {
				// Not first index
				buffer.WriteString(b[b_idxs[i-1]:b_idxs[i]])
			} else {
				// First index
				buffer.WriteString(b[:b_idxs[i]])
			}
		}
	}

	// If there were indicies we need to copy the last chunk from a into the buffer
	if last := len(b_idxs); last > 0 {
		buffer.WriteString(b[b_idxs[last-1]:])
	}

	return buffer.String(), nil
}

//
// Slice out the template tags
//
func tmpl_splice(tmpl string) string {
	// Collect % indicies
	idxs, err := lex(tmpl)
	if err != nil {
		log.Println(err)
	}

	// No tags to remove
	if len(idxs) == 0 {
		return tmpl
	}

	var buffer bytes.Buffer

	for i, v := range idxs {
		if i % 2 != 0 {
			_, value := extract(tmpl[idxs[i-1]:v])
			buffer.WriteString(value)
		} else {
			if i > 0 {
				buffer.WriteString(tmpl[idxs[i-1]:idxs[i]])
			} else {
				buffer.WriteString(tmpl[:idxs[0]])
			}
		}
	}

	if last := len(idxs); last > 0 {
		buffer.WriteString(tmpl[idxs[last-1]+1:])
	}

	return buffer.String()
}
