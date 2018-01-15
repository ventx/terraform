package plugin

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcErr filters unknown grpc and creates simple error strings from the
// message.
// Errors returned over grpc are converted to an internal type containing other
// protocol information. An unknown error type with no metadata is most likely
// an application error were only the message is relevant. Convert these to a
// simple error string for easier parsing by users.
func grpcErr(err error) error {
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok {
		// only contains a message, which means it's probably an application
		// error. Even if it's not, converting it to an error string loses no
		// information in this case.
		if s.Code() == codes.Unknown && len(s.Details()) == 0 {
			return errors.New(s.Message())
		}
	}
	return err
}
