package response

var Words map[string]string

func ResponseWord() {
	var word = make(map[string]string)
	word["enum"] = "Enum"
	word["value"] = "Value"
	word["order"] = "Order"
	word["match"] = "Match"

	Words = word
}
