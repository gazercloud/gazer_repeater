package fastspring

type ProductsGetAll struct {
	Action   string   `json:"action"`
	Result   string   `json:"result"`
	Products []string `json:"products"`
}

type AccountGetAll struct {
	Action   string   `json:"action"`
	Result   string   `json:"result"`
	Accounts []string `json:"accounts"`
}

type AccountGetItemContact struct {
	First   string `json:"first"`
	Last    string `json:"last"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
}

type AccountGetItemAddress struct {
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	City         string `json:"city"`
	PostalCode   string `json:"postalCode"`
	Region       string `json:"region"`
}

type AccountGetItem struct {
	Id      string                `json:"id"`
	Account string                `json:"account"`
	Action  string                `json:"action"`
	Contact AccountGetItemContact `json:"contact"`
	Address AccountGetItemAddress `json:"address"`
}

type AccountGet struct {
	Action   string           `json:"action"`
	Result   string           `json:"result"`
	Accounts []AccountGetItem `json:"accounts"`
}

type AccountCreate struct {
	Language string                `json:"language"`
	Country  string                `json:"country"`
	Contact  AccountGetItemContact `json:"contact"`
	Address  AccountGetItemAddress `json:"address"`
}

type AccountCreateResponse struct {
	Account string `json:"account"`
	Action  string `json:"action"`
	Result  string `json:"result"`
}

// gazercloud-pro

type SessionCreateItemPricingPrice struct {
	USD float64 `json:"USD"`
}

type SessionCreateItemPricing struct {
	Price SessionCreateItemPricingPrice `json:"price"`
}

type SessionCreateItem struct {
	Product  string `json:"product"`
	Quantity int64  `json:"quantity"`
	//Pricing SessionCreateItemPricing `json:"pricing"`
}

type SessionCreate struct {
	Account string              `json:"account"`
	Items   []SessionCreateItem `json:"items"`
}

type SessionCreateResponse struct {
	Id       string  `json:"id"`
	Expires  int64   `json:"expires"`
	Account  string  `json:"account"`
	SubTotal float64 `json:"subtotal"`
}
