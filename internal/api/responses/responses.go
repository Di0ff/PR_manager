package responses

type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Error Detail `json:"error"`
}

func Error(code string, message string) Response {
	return Response{
		Error: Detail{
			Code:    code,
			Message: message,
		},
	}
}
