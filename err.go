package proj

/* This module provides a special error type that is used for all errors
* returned by the memoizer */

import "fmt"

type MemErrCause int

const (
	MemErr_none       = iota //No Error
	MemErr_serErr            //An upstream service had an unrecoverable error
	MemErr_serCrash          //An upstream service crashed
	MemErr_serCorrupt        //An upstream service returned bad results
	MemErr_badArg            //Client sent an improper request
)

func (c MemErrCause) String() string {
	switch c {
	case MemErr_none:
		return "MemErr_none"
	case MemErr_serErr:
		return "MemErr_serErr"
	case MemErr_serCrash:
		return "MemErr_serCrash"
	case MemErr_serCorrupt:
		return "MemErr_serCorrupt"
	case MemErr_badArg:
		return "MemErr_badArg"
	default:
		return "MemErr_unknown"
	}
}

/* This is a special error type for the memoizer */
type MemErr struct {
	/* Why did this error happen? */
	cause MemErrCause
	/* An optional description of the error */
	desc string
	/* if cause=MemErr_serErr, this contains the upstream's error */
	serErr error
}

/* This allows MemErr to be printed out and generally treated like a standard error */
func (err MemErr) Error() string {
	switch err.cause {
	case MemErr_none:
		return fmt.Sprintf("No error to report (%s)", err.desc)
	case MemErr_serErr:
		return fmt.Sprintf("Upstream service had an error (%s): %v", err.desc, err.serErr)
	case MemErr_serCrash:
		return fmt.Sprintf("Upstream service failed (%s)", err.desc)
	case MemErr_serCorrupt:
		return fmt.Sprintf("Upstream service returned bad results (%s)", err.desc)
	case MemErr_badArg:
		return fmt.Sprintf("Bad client request (%s)", err.desc)
	default:
		return fmt.Sprintf("Unknown error %d (%s)", err.cause, err.desc)
	}
}

func CreateMemErr(cause MemErrCause, desc string, serErr error) MemErr {
	return MemErr{cause, desc, serErr}
}

func GetErrCause(e error) MemErrCause {
	return (e.(MemErr)).cause
}
