package session

import "io"

// closeFile closes c and stores the close error in *errp unless an earlier
// error is already there. Deferred on files opened for writing, it reports a
// failed flush on close instead of dropping it.
func closeFile(c io.Closer, errp *error) {
	if cerr := c.Close(); cerr != nil && *errp == nil {
		*errp = cerr
	}
}
