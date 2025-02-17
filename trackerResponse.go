package main

type TrackerResponseExpanded struct {
	FailureReason  string
	WarningMessage string
	Interval       int
	MinInterval    int
	TrackerID      []byte
	Complete       int
	Incomplete     int
	PeersDict      []Peer
	PeersBinary    []byte
}

type Peer struct {
	PeerID []byte
	IP     string
	Port   int
}

type TrackerResponseCompact struct {
	Interval int
	Peers    []byte
}

//GPT BUDDY

func UnpackTrackerResponse(data map[string]interface{}) TrackerResponseCompact {
	var response TrackerResponseCompact
	if value, ok := data["interval"].(int); ok {
		response.Interval = value
	}
	if peersBinary, ok := data["peers"].([]byte); ok {
		response.Peers = peersBinary
	}
	return response

	// var response TrackerResponseExpanded

	// if value, ok := data["failure reason"].([]byte); ok {
	// 	response.FailureReason = string(value)
	// }
	// if value, ok := data["warning message"].([]byte); ok {
	// 	response.WarningMessage = string(value)
	// }
	// if value, ok := data["interval"].(int); ok {
	// 	response.Interval = value
	// }
	// if value, ok := data["min interval"].(int); ok {
	// 	response.MinInterval = int(value)
	// }
	// if value, ok := data["tracker id"].([]byte); ok {
	// 	response.TrackerID = value
	// }
	// if value, ok := data["complete"].(int); ok {
	// 	response.Complete = value
	// }
	// if value, ok := data["incomplete"].(int); ok {
	// 	response.Incomplete = value
	// }

	// // Handle dictionary model peers
	// if peers, ok := data["peers"].([]interface{}); ok {
	// 	for _, peer := range peers {
	// 		if peerMap, ok := peer.(map[string]interface{}); ok {
	// 			var p Peer
	// 			if peerID, ok := peerMap["peer id"].([]byte); ok {
	// 				p.PeerID = peerID
	// 			}
	// 			if ip, ok := peerMap["ip"].([]byte); ok {
	// 				p.IP = ip
	// 			}
	// 			if port, ok := peerMap["port"].(int); ok {
	// 				p.Port = port
	// 			}
	// 			response.PeersDict = append(response.PeersDict, p)
	// 		}
	// 	}
	// }

	// // Handle binary model peers (assuming it's stored as a byte slice)
	// if peersBinary, ok := data["peers"].([]byte); ok {
	// 	response.PeersBinary = peersBinary
	// }

	// return response
}
