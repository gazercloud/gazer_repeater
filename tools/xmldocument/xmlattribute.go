package xmldocument

type XmlAttribute struct {
	name  string
	value string
}

func NewXmlAttribute() *XmlAttribute {
	var c XmlAttribute
	return &c
}

func (c *XmlAttribute) Name() string {
	return c.name
}

func (c *XmlAttribute) SetName(name string) {
	c.name = name
}

func (c *XmlAttribute) Value() string {
	return c.value
}

func (c *XmlAttribute) SetValue(value string) {
	c.value = value
}
