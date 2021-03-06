/* soup package implements a simple web scraper for Go,
keeping it as similar as possible to BeautifulSoup
*/

package soup

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Root is a structure containing a pointer to an html node, the node value, and an error variable to return an error if occurred
type Root struct {
	Pointer   *html.Node
	NodeValue string
	Error     error
}

// Debug status
// Setting this to true causes the panics to be thrown and logged onto the console.
// Setting this to false causes the errors to be saved in the Error field in the returned struct.
var Debug = false

// Headers contains all HTTP headers to send
var Headers = make(map[string]string)

// Cookies contains all HTTP cookies to send
var Cookies = make(map[string]string)

// GetWithClient returns the HTML returned by the url using a provided HTTP client
func GetWithClient(url string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if Debug {
			panic("Couldn't perform GET request to " + url)
		}
		return "", errors.New("couldn't perform GET request to " + url)
	}
	// Set headers
	for hName, hValue := range Headers {
		req.Header.Set(hName, hValue)
	}
	// Set cookies
	for cName, cValue := range Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cName,
			Value: cValue,
		})
	}
	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		if Debug {
			panic("Couldn't perform GET request to " + url)
		}
		return "", errors.New("couldn't perform GET request to " + url)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if Debug {
			panic("Unable to read the response body")
		}
		return "", errors.New("unable to read the response body")
	}
	return string(b), nil
}

// Get returns the HTML returned by the url in string using the default HTTP client
func Get(url string) (string, error) {
	// Init a new HTTP client
	client := &http.Client{}
	return GetWithClient(url, client)
}

// HTMLParse parses the HTML returning a start pointer to the DOM
func HTMLParse(s string) Root {
	r, err := html.Parse(strings.NewReader(s))
	if err != nil {
		if Debug {
			panic("Unable to parse the HTML")
		}
		return Root{nil, "", errors.New("unable to parse the HTML")}
	}
	for r.Type != html.ElementNode {
		switch r.Type {
		case html.DocumentNode:
			r = r.FirstChild
		case html.DoctypeNode:
			r = r.NextSibling
		case html.CommentNode:
			r = r.NextSibling
		}
	}
	return Root{r, r.Data, nil}
}

// Find finds the first occurrence of the given tag name,
// with or without attribute key and value specified,
// and returns a struct with a pointer to it
func (r Root) Find(name string, args ...string) Root {
	temp, ok := findOne(r.Pointer, name, args, false, false)
	if ok == false {
		if Debug {
			panic("Element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")
		}
		return Root{nil, "", errors.New("element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")}
	}
	return Root{temp, temp.Data, nil}
}

// FindAll finds all occurrences of the given tag name,
// with or without key and value specified,
// and returns an array of structs, each having
// the respective pointers
func (r Root) FindAll(name string, args ...string) []Root {
	temp := findAll(r.Pointer, name, args, false)
	if len(temp) == 0 {
		if Debug {
			panic("Element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")
		}
		return []Root{}
	}
	pointers := make([]Root, 0, len(temp))
	for i := 0; i < len(temp); i++ {
		pointers = append(pointers, Root{temp[i], temp[i].Data, nil})
	}
	return pointers
}

// FindStrict finds the first occurrence of the given tag name
// only if all the values of the provided attribute are an exact match
func (r Root) FindStrict(name string, args ...string) Root {
	temp, ok := findOne(r.Pointer, name, args, false, true)
	if ok == false {
		if Debug {
			panic("Element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")
		}
		return Root{nil, "", errors.New("element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")}
	}
	return Root{temp, temp.Data, nil}
}

// FindAllStrict finds all occurrences of the given tag name
// only if all the values of the provided attribute are an exact match
func (r Root) FindAllStrict(name string, args ...string) []Root {
	temp := findAll(r.Pointer, name, args, true)
	if len(temp) == 0 {
		if Debug {
			panic("Element `" + name + "` with attributes `" + strings.Join(args, " ") + "` not found")
		}
		return []Root{}
	}
	pointers := make([]Root, 0, len(temp))
	for i := 0; i < len(temp); i++ {
		pointers = append(pointers, Root{temp[i], temp[i].Data, nil})
	}
	return pointers
}

// FindNextSibling finds the next sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindNextSibling() Root {
	nextSibling := r.Pointer.NextSibling
	if nextSibling == nil {
		if Debug {
			panic("No next sibling found")
		}
		return Root{nil, "", errors.New("no next sibling found")}
	}
	return Root{nextSibling, nextSibling.Data, nil}
}

// FindPrevSibling finds the previous sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindPrevSibling() Root {
	prevSibling := r.Pointer.PrevSibling
	if prevSibling == nil {
		if Debug {
			panic("No previous sibling found")
		}
		return Root{nil, "", errors.New("no previous sibling found")}
	}
	return Root{prevSibling, prevSibling.Data, nil}
}

// FindNextElementSibling finds the next element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindNextElementSibling() Root {
	nextSibling := r.Pointer.NextSibling
	if nextSibling == nil {
		if Debug {
			panic("No next element sibling found")
		}
		return Root{nil, "", errors.New("no next element sibling found")}
	}
	if nextSibling.Type == html.ElementNode {
		return Root{nextSibling, nextSibling.Data, nil}
	}
	p := Root{nextSibling, nextSibling.Data, nil}
	return p.FindNextElementSibling()
}

// FindPrevElementSibling finds the previous element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindPrevElementSibling() Root {
	prevSibling := r.Pointer.PrevSibling
	if prevSibling == nil {
		if Debug {
			panic("No previous element sibling found")
		}
		return Root{nil, "", errors.New("no previous element sibling found")}
	}
	if prevSibling.Type == html.ElementNode {
		return Root{prevSibling, prevSibling.Data, nil}
	}
	p := Root{prevSibling, prevSibling.Data, nil}
	return p.FindPrevElementSibling()
}

// Children returns all direct children of this DOME element.
func (r Root) Children() []Root {
	var cs []Root
	for c := r.Pointer.FirstChild; c != nil; c = c.NextSibling {
		cs = append(cs, Root{c, c.Data, nil})
	}
	return cs
}

// Attrs returns a map containing all attributes
func (r Root) Attrs() map[string]string {
	if r.Pointer.Type != html.ElementNode {
		if Debug {
			panic("Not an ElementNode")
		}
		return nil
	}
	if len(r.Pointer.Attr) == 0 {
		return nil
	}
	return getKeyValue(r.Pointer.Attr)
}

// Text returns the string inside a non-nested element
func (r Root) Text() string {
	k := r.Pointer.FirstChild
checkNode:
	if k != nil && k.Type != html.TextNode {
		k = k.NextSibling
		if k == nil {
			if Debug {
				panic("No text node found")
			}
			return ""
		}
		goto checkNode
	}
	if k != nil {
		r, _ := regexp.Compile(`^\s+$`)
		if ok := r.MatchString(k.Data); ok {
			k = k.NextSibling
			if k == nil {
				if Debug {
					panic("No text node found")
				}
				return ""
			}
			goto checkNode
		}
		return k.Data
	}
	return ""
}

// FullText returns the string inside even a nested element
func (r Root) FullText() string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		if n.Type == html.ElementNode {
			f(n.FirstChild)
		}
		if n.NextSibling != nil {
			f(n.NextSibling)
		}
	}

	f(r.Pointer.FirstChild)

	return buf.String()
}

// Using depth first search to find the first occurrence and return
func findOne(n *html.Node, name string, args []string, uni bool, strict bool) (*html.Node, bool) {
	if uni == true {
		if n.Type == html.ElementNode && n.Data == name {
			if len(args) == 2 {
				for i := 0; i < len(n.Attr); i++ {
					attr := n.Attr[i]
					searchAttrName := args[0]
					searchAttrVal := args[1]
					if (strict && attributeAndValueEquals(attr, searchAttrName, searchAttrVal)) ||
						(!strict && attributeContainsValue(attr, searchAttrName, searchAttrVal)) {
						return n, true
					}
				}
			} else if len(args) == 0 {
				return n, true
			}
		}
	}
	uni = true
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p, q := findOne(c, name, args, true, strict)
		if q != false {
			return p, q
		}
	}
	return nil, false
}

// Using depth first search to find all occurrences and return
func findAll(n *html.Node, name string, args []string, strict bool) []*html.Node {
	var nodeLinks = make([]*html.Node, 0, 10)
	var f func(*html.Node, []string, bool)
	f = func(n *html.Node, args []string, uni bool) {
		if uni == true {
			if n.Data == name {
				if len(args) == 2 {
					for i := 0; i < len(n.Attr); i++ {
						attr := n.Attr[i]
						searchAttrName := args[0]
						searchAttrVal := args[1]
						if (strict && attributeAndValueEquals(attr, searchAttrName, searchAttrVal)) ||
							(!strict && attributeContainsValue(attr, searchAttrName, searchAttrVal)) {
							nodeLinks = append(nodeLinks, n)
						}
					}
				} else if len(args) == 0 {
					nodeLinks = append(nodeLinks, n)
				}
			}
		}
		uni = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, args, true)
		}
	}
	f(n, args, false)
	return nodeLinks
}

// attributeAndValueEquals reports when the html.Attribute attr has the same attribute name and value as from
// provided arguments
func attributeAndValueEquals(attr html.Attribute, attribute, value string) bool {
	return attr.Key == attribute && attr.Val == value
}

// attributeContainsValue reports when the html.Attribute attr has the same attribute name as from provided
// attribute argument and compares if it has the same value in its values parameter
func attributeContainsValue(attr html.Attribute, attribute, value string) bool {
	if attr.Key == attribute {
		for _, attrVal := range strings.Fields(attr.Val) {
			if attrVal == value {
				return true
			}
		}
	}
	return false
}

// Returns a key pair value for each attribute
func getKeyValue(attributes []html.Attribute) map[string]string {
	var keyValues = make(map[string]string)
	for _, a := range attributes {
		if _, exists := keyValues[a.Key]; !exists {
			keyValues[a.Key] = a.Val
		}
	}
	return keyValues
}
