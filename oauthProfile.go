package main

/*
	Google profile struct used to parse json response
*/
type ProfileBlob struct {
	Emails        []Emails        `json:"emailAddresses"`
	Names         []Name          `json:"names"`
	Images        []Image         `json:"photos"`
	Organizations []Organizations `json:"organizations"`
}

/*
	google profile returns array of emails based on number of gmail sessions user is logged in as
	So this helps capture them all so we can later parse for address
*/
type Emails struct {
	Value     string `json:"value"`
	EmailType string `json:"type"`
}

/*
	Name is return as an object so use this struct to parse the name fields
*/
type Name struct {
	displayName string `json:"displayName`
	FamilyName  string `json:"familyName"`
	GivenName   string `json:"givenName"`
}

/*
	Parse image object in response
*/
type Image struct {
	URL       string `json:"url"`
	IsDefault bool   `json:"default"`
}

/*
	parse organizations object in response.  Google will return an array of orgs
*/
type Organizations struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}
