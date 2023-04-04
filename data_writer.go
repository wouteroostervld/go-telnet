package telnet

import (
	"bufio"
	"io"
	"log"
)

// An internalDataWriter deals with "escaping" according to the TELNET (and TELNETS) protocol.
//
// In the TELNET (and TELNETS) protocol byte value 255 is special.
//
// The TELNET (and TELNETS) protocol calls byte value 255: "IAC". Which is short for "interpret as command".
//
// The TELNET (and TELNETS) protocol also has a distinction between 'data' and 'commands'.
//
// (DataWriter is targetted toward TELNET (and TELNETS) 'data', not TELNET (and TELNETS) 'commands'.)
//
// If a byte with value 255 (=IAC) appears in the data, then it must be escaped.
//
// Escaping byte value 255 (=IAC) in the data is done by putting 2 of them in a row.
//
// So, for example:
//
//	[]byte{255} -> []byte{255, 255}
//
// Or, for a more complete example, if we started with the following:
//
//	[]byte{1, 55, 2, 155, 3, 255, 4, 40, 255, 30, 20}
//
// ... TELNET escaping would produce the following:
//
//	[]byte{1, 55, 2, 155, 3, 255, 255, 4, 40, 255, 255, 30, 20}
//
// (Notice that each "255" in the original byte array became 2 "255"s in a row.)
//
// internalDataWriter takes care of all this for you, so you do not have to do it.
type internalDataWriter struct {
	wrapped *bufio.Writer
}

// newDataWriter creates a new internalDataWriter writing to 'w'.
//
// 'w' receives what is written to the *internalDataWriter but escaped according to
// the TELNET (and TELNETS) protocol.
//
// I.e., byte 255 (= IAC) gets encoded as 255, 255.
//
// For example, if the following it written to the *internalDataWriter's Write method:
//
//	[]byte{1, 55, 2, 155, 3, 255, 4, 40, 255, 30, 20}
//
// ... then (conceptually) the following is written to 'w's Write method:
//
//	[]byte{1, 55, 2, 155, 3, 255, 255, 4, 40, 255, 255, 30, 20}
//
// (Notice that each "255" in the original byte array became 2 "255"s in a row.)
//
// *internalDataWriter takes care of all this for you, so you do not have to do it.
func newDataWriter(w io.Writer) *internalDataWriter {
	b := bufio.NewWriter(w)
	return &internalDataWriter{wrapped: b}
}

// Write writes the TELNET (and TELNETS) escaped data for of the data in 'data' to the wrapped io.Writer.
func (w *internalDataWriter) Write(data []byte) (n int, err error) {

	// loop through the data, looking for IACs
	// if we find one, write another one
	// flush the buffer

	var n_total int = 0
	for i := 0; i < len(data); i++ {
		if data[i] == 255 {
			log.Printf(("Found IAC at %d"), i))
			// we found an IAC
			// write the buffer up to this point
			// write the IAC
			n, e := w.wrapped.Write(data[:i])
			n_total += n
			if e != nil {
				log.Printf("Flushing")
				w.wrapped.Flush()
				return n_total, e
			}
			e = w.wrapped.WriteByte(255)
			if e != nil {
				return n_total, e
			}
			log.Printf("Flushing")
			w.wrapped.Flush()
			n_total += 1
			e = w.wrapped.WriteByte(255)
			if e != nil {
				log.Printf("Flushing")
				w.wrapped.Flush()
				return n_total, e
			}
			data = data[i+1:]
			i = 0
		}
	}
	n, e := w.wrapped.Write(data)
	n_total += n
	if e != nil {
		log.Printf("Flushing")
		w.wrapped.Flush()
		return n_total, e
	}
	log.Printf("Flushing")
	w.wrapped.Flush()
	return n_total, nil
}
