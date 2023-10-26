package response

var Message map[string]string

func ResponseMsg() {
	var message = make(map[string]string)
	message["went_wrong"] = "Something went wrong"
	message["server_error"] = "Internal server error"
	message["not_match"] = "dose not match"
	message["invalid"] = "validation failed"
	message["match_found"] = "Match found successful"

	Message = message
}
