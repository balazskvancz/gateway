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

	errServicesIsNil            = errors.New("services is nil")
	errServicesPrefixLength     = errors.New("service prefix must be greater than zero")
	errServicesSamePrefixLength = errors.New("service prefix must be same length")
	errServicesSliceIsEmpty     = errors.New("services slice is empty")

	errNoService      = errors.New("zero length of services")
	errRegistryNil    = errors.New("[registry]: registry is nil")
	errServiceExists  = errors.New("[registry]: service already registered")
	errServiceTreeNil = errors.New("[registry]: service tree is <nil>")
	errServiceNil     = errors.New("[registry]: service is nil")

	ErrServiceNotExists = errors.New("service not exists")

	errNotJsonContentType = errors.New("[context]: incoming body is not application/json")
	errDataMustBePtr      = errors.New("[context]: data must be a pointer")
)
