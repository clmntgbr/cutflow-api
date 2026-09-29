package mediafile

import "errors"

var (
	ErrInvalidFilename   = errors.New("invalid media filename")
	ErrUnsupportedType   = errors.New("unsupported media type")
	ErrInvalidTransition = errors.New("invalid media status transition")
	ErrMediaNotFound     = errors.New("media file not found")
	ErrMediaTooLarge     = errors.New("media exceeds maximum size")
	ErrProbeInvalid      = errors.New("media file is not a usable video")
)
