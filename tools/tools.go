package tools

func UpdateListener(){
	service := "0.0.0.0:1111"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		utils.Log("Error: Could not resolve address ", err)
	} else {
		netListen, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if err != nil {
			utils.Log(err)
		} else {
			defer netListen.Close()
			for {
				conn, err := netListen.Accept()
				if err != nil {
					utils.Log("Client error: ", err)
				} else {
					go UpdateHandler(conn)
				}
			}
		}
	}

}

func UpdateHandler(){
	jobj, err := utils.ReadJSON(conn)
	if err == nil {
		subId, err := Login(jobj, conn)
		if err == nil {
			jobj, err = utils.ReadJSON(conn)
			if err == nil {
				req, err := utils.JSONValue(jobj, REQ)
				if req == SEND {
					err = ProcessFile(subId, jobj, conn, procChan)
				} else if req == LOGOUT {
					break
				} else if err == nil {
					err = errors.New("Unknown request: " + req)
				}
			}
		}
	}
	EndSession(conn, err)




}