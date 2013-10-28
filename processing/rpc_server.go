package processing

import (
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type (
	RPCHandler struct {
		sub    chan *SubRequest
		file   chan *FileRequest
		status chan Status
	}
)

func setupRPC(port int, subCh chan *SubRequest, fileCh chan *FileRequest, statusCh chan Status) (err error) {
	handler := &RPCHandler{
		sub:    subCh,
		file:   fileCh,
		status: statusCh,
	}
	rpc.Register(handler)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err == nil {
		go http.Serve(l, nil)
	}
	return
}

//GetStatus retrieves Impendulo's current processing status.
func (this *RPCHandler) GetStatus(nothing Empty, ret *Status) error {
	this.status <- Status{0, 0}
	s := <-this.status
	ret.Files, ret.Submissions = s.Files, s.Submissions
	return nil
}

//AddFile sends a file id to be processed.
func (this *RPCHandler) AddFile(req *FileRequest, nothing *Empty) error {
	req.Response = make(chan error)
	//We send the file's db id as well as the id of the submission to which it belongs.
	this.file <- req
	//Return any errors which occured while adding the file.
	return <-req.Response
}

//StartSubmission signals that this submission will now receive files.
func (this *RPCHandler) ToggleSubmission(req *SubRequest, nothing *Empty) error {
	req.Response = make(chan error)
	this.sub <- req
	return <-req.Response
}
