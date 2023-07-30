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

	errBadProtocol = errors.New("[service]: only http or https protocol are supported")
	errConfigIsNil = errors.New("[service]: config is <nil>")
	errEmptyHost   = errors.New("[service]: hostname cant be empty")
	errEmptyName   = errors.New("[service]: name cant be empty")
	errEmptyPort   = errors.New("[service]: port cant be empty")
	errEmptyPrefix = errors.New("[service]: prefix cant be empty")

	errServiceNotAvailable = errors.New("service is not available")

	errRegistryNil      = errors.New("[registry]: registry is nil")
	errServiceExists    = errors.New("[registry]: service already registered")
	errServiceTreeNil   = errors.New("[registry]: service tree is <nil>")
	ErrServiceNotExists = errors.New("[registry]: service not exists")

	errNotJsonContentType = errors.New("[context]: incoming body is not application/json")
	errDataMustBePtr      = errors.New("[context]: data must be a pointer")

	errContextIsNil             = errors.New("given context is nil")
	errKeyInContextIsNotPresent = errors.New("the key is not stored in the given context")
)
