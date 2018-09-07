package errno

var (
	// Common errors
	OK                  = &Errno{Code: 200, Message: "Success"}
	InternalServerError = &Errno{Code: 10001, Message: "Internal server error."}
	ErrBind             = &Errno{Code: 10002, Message: "Error occurred while binding the request body to the struct."}
	ParamsErr           = &Errno{Code: 10003, Message: "Missing param: "}

	DbError = &Errno{Code: 30100, Message: "The database error."}
)
