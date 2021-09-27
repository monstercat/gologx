package logx

type GoogleLogHandler struct {

}

func (h *GoogleLogHandler) Handle(l Log) (int, error) {
	return 0, nil
}