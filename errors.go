package gateway

import "errors"

var (
	// Tree
	errBadPathParamSyntax = errors.New("[tree]: bad path param syntax")
	errKeyIsAlreadyStored = errors.New("[tree]: key is already stored")
	errKeyIsEmpty         = errors.New("[tree]: key is empty")
	errMissingSlashPrefix = errors.New("[tree]: urls must be started with a '/'")
	errNoCommonPrefix     = errors.New("[tree]: no commmon prefix in given strings")
	errPresentSlashSuffix = errors.New("[tree]: urls must not be ended with a '/'")
	errRootIsNil          = errors.New("[tree]: the root of the tree is <nil>")
	errTreeIsNil          = errors.New("[tree]: the tree is <nil>")

	errServiceNotAvailable = errors.New("service is not available")

	errParamNotExists = errors.New("[context]: requested param not exists")

	errServicesIsNil            = errors.New("services is nil")
	errServicesPrefixLength     = errors.New("service prefix must be greater than zero")
	errServicesSamePrefixLength = errors.New("service prefix must be same length")
	errServicesSliceIsEmpty     = errors.New("services slice is empty")

	errNoService     = errors.New("zero length of services")
	errRegistryNil   = errors.New("registry is nil")
	errServiceExists = errors.New("service already registered")
	errServiceMapNil = errors.New("service registry is nil")
	errServiceNil    = errors.New("service is nil")

	ErrServiceNotExists = errors.New("service not exists")
)
