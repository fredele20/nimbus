package service

var (
	ErrWalletNotFound          = new("wallet not found")
	ErrWalletAlreadyExists     = new("wallet already exists for this currency")
	ErrInvalidCurrency         = new("unsupported wallet currency")
	ErrInvalidStatus           = new("invalid wallet status")
	ErrInvalidStatusTransition = new("invalid wallet status transition")
	ErrInvalidUserID           = new("invalid user ID")
	ErrInvalidWalletID         = new("invalid wallet ID")

	ErrWalletRestrictionConflict = new("wallet restriction conflict")
	ErrWalletAccessDenied        = new("wallet access denied")
)

type walletErr struct {
	message string
}

func (w *walletErr) Error() string {
	return w.message
}

func new(msg string) *walletErr {
	return &walletErr{message: msg}
}
