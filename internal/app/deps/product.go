package deps

import (
	"context"

	"github.com/assanoff/servicekit/dim"

	productapi "github.com/assanoff/service-kit-x/api/product"
	"github.com/assanoff/service-kit-x/core/product"
	"github.com/assanoff/service-kit-x/core/product/productdb"
)

// initProductCore builds the product business core over the Postgres store.
var initProductCore = func(c *Deps) (dim.CleanupFunc, error) {
	c.ProductCore = dim.OnceWithName("ProductCore", func(ctx context.Context) (*product.Core, error) {
		return product.NewCore(c.Logger, productdb.NewStore(c.Logger, c.DB(ctx))), nil
	})
	return nil, nil
}

// initProductHandler builds the REST handler for products.
var initProductHandler = func(c *Deps) (dim.CleanupFunc, error) {
	c.ProductHandler = dim.OnceWithName("ProductHandler", func(ctx context.Context) (*productapi.Handler, error) {
		return productapi.New(c.ProductCore(ctx)), nil
	})
	return nil, nil
}
