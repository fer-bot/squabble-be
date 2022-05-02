package form

type AnswerRequest struct {
	Word string `json:"word"`
}

type AnswerResponse struct {
	Mark [5]string `json:"mark"`
}

func AnswerResponseBuilder(mark [5]string) *AnswerResponse {
	return &AnswerResponse{mark}
}
