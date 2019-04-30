package util

import "io"

// WriteWithGuarantee sends a message to the given writer
// The message is guaranteed to be fully delivered, or the function returns with an error
func WriteWithGuarantee(conn io.Writer, message []byte) error {

	written := 0

	for written < len(message) {
		wrote, err := conn.Write(message[written:])
		if err != nil {
			return err
		}

		written += wrote
	}

	return nil
}
