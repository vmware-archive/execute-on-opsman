package client

type OpsManClient interface {
	// Products()
	// Manifest(product_id string)
	// VMCredentials(product_id string)
	// VMNames(product_id string)
	DirectorManifest()
	// ProductId(product_type string)
}
