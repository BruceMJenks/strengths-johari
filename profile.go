package main


/*
	Google profile struct used to parse json response
*/
type ProfileBlob struct {
		Kind string
		Etag string
		Gender string
		Emails []Emails
		ObjectType string
		Id string
		DisplayName string
		Name Name
		Url string
		Image Image
		Organizations []Organizations
		PlacesLived []PlacesLived
		IsPlusUser bool
		Language string
		CircledByCount int
		Verified bool
		Domain string
}

/*
	google profile returns array of emails based on number of gmail sessions user is logged in as
	So this helps capture them all so we can later parse for address
*/
type Emails struct {
	Value string
	Typee string
}

/*
	Name is return as an object so use this struct to parse the name fields
*/
type Name struct {
	FamilyName 	string
	GivenName 	string
}

/*
	Parse image object in response
*/
type Image struct {
	Url string
	IsDefault bool
}

/*
	parse organizations object in response.  Google will return an array of orgs
*/
type Organizations struct{
	Name string
	Title string
	EndDate string
	Primary bool
}

/*
	google returns array of addresses so need this struct to parse them all
*/
type PlacesLived struct{
	Value string
	Primary bool
}
