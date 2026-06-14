package internal

import "crypto/rand"

func PrepareData(json []byte) (data []byte) {
	data = append(data, 0x81)
	if len(json) < 126 {
		data = append(data, 0x80+byte(len(json)))
	} else if len(json) < 65536 {
		data = append(data, 0x80+126)
		data = append(data, byte(len(json)>>8), byte(len(json)))
	} else {
		data = append(data, 0x80+127)
	}

	key := make([]byte, 4)
	rand.Reader.Read(key)
	data = append(data, key...)

	for i := range json {
		data = append(data, byte(json[i]^key[i%4]))
	}

	return
}
