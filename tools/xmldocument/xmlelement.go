package xmldocument

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

type XmlElement struct {
	name  string
	attrs []*XmlAttribute
	text  string

	elements []*XmlElement
}

func NewXmlElement() *XmlElement {
	var c XmlElement
	c.elements = make([]*XmlElement, 0)
	c.attrs = make([]*XmlAttribute, 0)
	return &c
}

func (c *XmlElement) AddElement(name string) *XmlElement {
	var el XmlElement
	el.name = name
	elPointer := &el
	c.elements = append(c.elements, elPointer)
	return elPointer
}

func (c *XmlElement) SetAttribute(name string, value string) *XmlAttribute {
	var attr XmlAttribute
	attr.name = name
	attr.value = value
	attrPointer := &attr
	c.attrs = append(c.attrs, attrPointer)
	return attrPointer
}

func (c *XmlElement) Attributes() []*XmlAttribute {
	return c.attrs
}

func (c *XmlElement) Elements() []*XmlElement {
	return c.elements
}

func (c *XmlElement) Name() string {
	return c.name
}

func (c *XmlElement) SetName(name string) {
	c.name = name
}

func (c *XmlElement) FindFirstElementByName(name string) (*XmlElement, error) {
	for _, a := range c.elements {
		if a.name == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

func (c *XmlElement) FindAllElementsByName(name string) []*XmlElement {
	var result []*XmlElement
	for _, a := range c.elements {
		if a.name == name {
			result = append(result, a)
		}
	}
	return result
}

func (c *XmlElement) FindAttributeByName(name string) (*XmlAttribute, error) {
	for _, a := range c.attrs {
		if a.name == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

func (c *XmlElement) AttrValueInt8(name string, defaultValue int8) int8 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseInt(a.value, 10, 8)
			if err != nil {
				return defaultValue
			}
			return int8(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueUInt8(name string, defaultValue uint8) uint8 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseUint(a.value, 10, 8)
			if err != nil {
				return defaultValue
			}
			return uint8(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueInt16(name string, defaultValue int16) int16 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseInt(a.value, 10, 16)
			if err != nil {
				return defaultValue
			}
			return int16(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueUInt16(name string, defaultValue uint16) uint16 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseUint(a.value, 10, 16)
			if err != nil {
				return defaultValue
			}
			return uint16(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueInt32(name string, defaultValue int32) int32 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseInt(a.value, 10, 32)
			if err != nil {
				return defaultValue
			}
			return int32(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueUInt32(name string, defaultValue uint32) uint32 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseUint(a.value, 10, 32)
			if err != nil {
				return defaultValue
			}
			return uint32(v)
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueInt64(name string, defaultValue int64) int64 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseInt(a.value, 10, 64)
			if err != nil {
				return defaultValue
			}
			return v
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueUInt64(name string, defaultValue uint64) uint64 {
	for _, a := range c.attrs {
		if a.name == name {
			v, err := strconv.ParseUint(a.value, 10, 64)
			if err != nil {
				return defaultValue
			}
			return v
		}
	}
	return defaultValue
}

func (c *XmlElement) AttrValueString(name string, defaultValue string) string {
	for _, a := range c.attrs {
		if a.name == name {
			return a.value
		}
	}
	return defaultValue
}

func (c *XmlElement) saveToEncoder(encoder *xml.Encoder) error {
	var se xml.StartElement
	se.Name.Local = c.Name()
	se.Attr = make([]xml.Attr, 0)

	for _, attr := range c.attrs {
		var a xml.Attr
		a.Name.Local = attr.name
		a.Value = attr.value
		se.Attr = append(se.Attr, a)
	}

	err := encoder.EncodeToken(se)
	if err != nil {
		return err
	}

	for _, el := range c.elements {
		if err = el.saveToEncoder(encoder); err != nil {
			return err
		}
	}

	var ee xml.EndElement
	ee.Name.Local = c.name
	err = encoder.EncodeToken(ee)
	if err != nil {
		return err
	}

	return nil
}
