package xmldocument

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/golang-collections/collections/stack"
	"io"
	"io/ioutil"
	"os"
)

type XmlDocument struct {
	rootElement *XmlElement
	err         error
}

func NewXmlDocument() *XmlDocument {
	var c XmlDocument
	return &c
}

func (c *XmlDocument) CreateRootElement(name string) *XmlElement {
	var rootEl XmlElement
	rootEl.name = name
	rootElPointer := &rootEl
	c.rootElement = rootElPointer
	return rootElPointer
}

func (c *XmlDocument) RootElement() *XmlElement {
	return c.rootElement
}

func (c *XmlDocument) LoadFromFile(fileName string) error {
	var err error
	bytesResult, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(bytesResult)

	decoder := xml.NewDecoder(buf)

	var elStack stack.Stack

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			if c.rootElement == nil {
				return fmt.Errorf("no root item found")
			}
			err = nil
			break
		}

		if err != nil {
			return err
		}

		if se, is := t.(xml.StartElement); is {
			el := NewXmlElement()
			el.name = se.Name.Local

			for _, a := range se.Attr {
				attr := NewXmlAttribute()
				attr.name = a.Name.Local
				attr.value = a.Value
				el.attrs = append(el.attrs, attr)
			}

			if elStack.Len() > 0 {
				elStack.Peek().(*XmlElement).elements = append(elStack.Peek().(*XmlElement).elements, el)
			}
			elStack.Push(el)
		}

		if cd, is := t.(xml.CharData); is {
			if elStack.Len() > 0 {
				elStack.Peek().(*XmlElement).text += string(cd.Copy())
			} else {
				return fmt.Errorf("no tag found for char data")
			}
		}

		if ee, is := t.(xml.EndElement); is {
			if elStack.Len() < 1 {
				return fmt.Errorf("no start tag found")
			}
			el := elStack.Pop()
			if ee.Name.Local != el.(*XmlElement).name {
				return fmt.Errorf("wrong tag name")
			}
			if elStack.Len() == 0 {
				c.rootElement = el.(*XmlElement)
			}
		}
	}

	return err
}

func (c *XmlDocument) SaveToFile(fileName string) error {
	if c.rootElement == nil {
		return fmt.Errorf("no root element found")
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	encoder := xml.NewEncoder(buf)
	encoder.Indent("", " ")

	err := c.rootElement.saveToEncoder(encoder)
	if err != nil {
		return err
	}

	if err = encoder.Flush(); err != nil {
		return err
	}

	if err := ioutil.WriteFile(fileName, buf.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return err
}
