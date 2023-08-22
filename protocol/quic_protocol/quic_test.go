package quic_protocol

// func TestQuic(t *testing.T) {
// 	quic := &QuicProtocol{}
// 	wait := sync.WaitGroup{}
// 	wait.Add(1)
// 	go quic.ListenAndServe(80, func(connection protocol.Connection) {
// 		for {
// 			msg, err := connection.ReadMsg()
// 			if err != nil {
// 				break
// 			}
// 			fmt.Printf("%s\n", string(msg))
// 			wait.Done()
// 		}
// 	})
// 	connection, err := quic.Dial("127.0.0.1:80")
// 	if err != nil {
// 		fmt.Printf("%#v\n", err.Error())
// 	}
// 	err = connection.WriteMsg([]byte("你好"))
// 	if err != nil {
// 		fmt.Printf("%#v\n", err.Error())
// 	}
// 	wait.Wait()
// }
